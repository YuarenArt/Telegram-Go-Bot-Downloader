package tg

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
	"youtube_downloader/pkg/bot/tg/send"
	"youtube_downloader/pkg/database/models"
)

const (
	commandStart  = "start"
	commandHelp   = "help"
	commandPay    = "pay"
	commandStatus = "status"

	payMonth    = "pay_month"
	payYear     = "pay_year"
	payLifetime = "pay_lifetime"
)

// handleCommand handles supported commands
func (tb *TgBot) handleCommand(message *tgbotapi.Message) {
	lang := message.From.LanguageCode
	switch message.Command() {
	case commandStart:
		tb.handleStartCommand(message, lang)
	case commandHelp:
		tb.handleHelpCommand(message, lang)
	case commandPay:
		tb.handlePayCommand(message)
	case commandStatus:
		tb.UserStatus(message, lang)
	default:
		tb.handleDefaultCommand(message, lang)
	}
}

// handlePayCommand handles the /pay command with or without a subscription type
func (tb *TgBot) handlePayCommand(message *tgbotapi.Message) {
	subscriptionType := message.CommandArguments()

	if subscriptionType == "" {
		tb.sendPayOptions(message)
		return
	}

	tb.processPayment(message, subscriptionType)
}

// sendPayOptions sends buttons with subscription options to the user
func (tb *TgBot) sendPayOptions(message *tgbotapi.Message) {
	lang := message.From.LanguageCode
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(tb.translations[lang]["monthlyButton"], payMonth),
		tgbotapi.NewInlineKeyboardButtonData(tb.translations[lang]["yearlyButton"], payYear),
		tgbotapi.NewInlineKeyboardButtonData(tb.translations[lang]["lifetimeButton"], payLifetime),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	msg := tgbotapi.NewMessage(message.Chat.ID, tb.translations[lang]["chooseSubscriptionPlan"])
	msg.ReplyMarkup = keyboard

	if _, err := tb.Bot.Send(msg); err != nil {
		log.Println("Error sending payment options:", err)
	}
}

// processPayment processes the payment based on the selected subscription type
func (tb *TgBot) processPayment(message *tgbotapi.Message, subscriptionType string) {
	lang := message.From.LanguageCode
	if tb.Client == nil {
		errMsg := "Database client is not initialized"
		log.Println(errMsg)
		send.SendReplyMessage(tb.Bot, message, &errMsg)
		return
	}

	switch subscriptionType {
	case payMonth, payYear, payLifetime:
		tb.sendInvoice(message, subscriptionType, lang)
	default:
		tb.sendPayOptions(message)
	}
}

// sendInvoice sends an invoice to the user
func (tb *TgBot) sendInvoice(message *tgbotapi.Message, subscriptionType string, lang string) {
	subscriptions := map[string]struct {
		Title       string
		Description string
		Amount      int
	}{
		"month": {
			Title:       tb.translations[lang]["monthlyTitle"],
			Description: tb.translations[lang]["monthlyDescription"],
			Amount:      10000,
		},
		"year": {
			Title:       tb.translations[lang]["yearlyTitle"],
			Description: tb.translations[lang]["yearlyDescription"],
			Amount:      100000,
		},
		"lifetime": {
			Title:       tb.translations[lang]["lifetimeTitle"],
			Description: tb.translations[lang]["lifetimeDescription"],
			Amount:      200000,
		},
	}

	subscription, exists := subscriptions[subscriptionType]
	if !exists {
		send.SendMessage(tb.Bot, message, tb.translations[lang]["invalidSubscriptionType"])
		return
	}

	prices := []tgbotapi.LabeledPrice{
		{
			Label:  subscription.Title,
			Amount: subscription.Amount,
		},
	}
	payload := subscriptionType

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading env.example file: %v", err)
	}
	providerToken := os.Getenv("PROVIDER_TOKEN")
	if providerToken == "" {
		log.Fatal("Can't get provider token")
	}

	invoice := tgbotapi.NewInvoice(
		message.Chat.ID,
		subscription.Title,
		subscription.Description,
		payload,
		providerToken,
		"",
		"RUB",
		prices,
	)

	invoice.SuggestedTipAmounts = []int{}

	if _, err := tb.Bot.Request(invoice); err != nil {
		log.Println("Error sending invoice:", err)
	}
}

func (tb *TgBot) handleSuccessfulPayment(message *tgbotapi.Message) {
	if tb.Client == nil {
		errMsg := "Database client is not initialized"
		log.Println(errMsg)
		send.SendMessage(tb.Bot, message, errMsg)
		return
	}

	log.Printf("Successful payment from %s, amount: %d", message.From.UserName, message.SuccessfulPayment.TotalAmount)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	username := message.From.UserName
	if username == "" {
		username = fmt.Sprintf("user_%d", message.From.ID)
	}

	// Get current subscription status
	status, err := tb.Client.GetSubscriptionStatus(ctx, username)
	if err != nil {
		log.Printf("Error getting subscription status: %v", err)
		send.SendMessage(tb.Bot, message, tb.translations[message.From.LanguageCode]["errorGettingSubscription"])
		return
	}

	// Check if the subscription is still active
	now := time.Now()
	isActive := now.Before(status.EndSubscription)
	if isActive {
		log.Printf("User %s already has an active subscription", username)
		send.SendMessage(tb.Bot, message, tb.translations[message.From.LanguageCode]["subscriptionAlreadyActive"])
		return
	}

	// Calculate new subscription end time
	endTime := now
	switch message.SuccessfulPayment.InvoicePayload {
	case payMonth:
		endTime = now.AddDate(0, 1, 0)
	case payYear:
		endTime = now.AddDate(1, 0, 0)
	case payLifetime:
		endTime = now.AddDate(900, 0, 0)
	}

	// Update user subscription
	updateUser := &models.User{
		Username: username,
		ChatID:   message.From.ID,
		Subscription: models.Subscription{
			Duration:          message.SuccessfulPayment.InvoicePayload,
			StartSubscription: now,
			EndSubscription:   endTime,
		},
	}

	err = tb.Client.UpdateSubscription(ctx, updateUser)
	if err != nil {
		log.Printf("Error updating user subscription: %s", err.Error())
		send.SendMessage(tb.Bot, message, tb.translations[message.From.LanguageCode]["errorUpdatingSubscription"])
		return
	}

	send.SendMessage(tb.Bot, message, tb.translations[message.From.LanguageCode]["thankYouForPayment"])
}

// handleStartCommand sends a message with startMessage text
func (tb *TgBot) handleStartCommand(message *tgbotapi.Message, lang string) error {
	return send.SendMessage(tb.Bot, message, tb.translations[lang]["startMessage"])
}

// handleHelpCommand sends a message with helpMessage text
func (tb *TgBot) handleHelpCommand(message *tgbotapi.Message, lang string) error {
	return send.SendMessage(tb.Bot, message, tb.translations[lang]["helpMessage"])
}

// handleDefaultCommand sends a message with defaultMessage text
func (tb *TgBot) handleDefaultCommand(message *tgbotapi.Message, lang string) error {
	return send.SendMessage(tb.Bot, message, tb.translations[lang]["defaultMessage"])
}

// UserStatus sends the user's subscription status and subscription expiration date if active
func (tb *TgBot) UserStatus(message *tgbotapi.Message, lang string) error {
	if tb.Client == nil {
		errMsg := "Database client is not initialized"
		log.Println(errMsg)
		send.SendReplyMessage(tb.Bot, message, &errMsg)
		return fmt.Errorf(errMsg)
	}

	if message == nil || message.From == nil {
		errMsg := "invalid message or user information"
		send.SendReplyMessage(tb.Bot, message, &errMsg)
		return fmt.Errorf(errMsg)
	}

	username := message.From.UserName
	if username == "" {
		username = fmt.Sprintf("user_%d", message.From.ID)
	}

	status, err := tb.Client.GetSubscriptionStatus(context.Background(), username)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting subscription status: %v", err)
		send.SendReplyMessage(tb.Bot, message, &errMsg)
		return fmt.Errorf("error getting subscription status: %w", err)
	}

	now := time.Now()
	isActive := now.Before(status.EndSubscription)

	var statusText string
	if isActive {
		expiresAt := status.EndSubscription.Format("2006-01-02 15:04:05")
		statusText = fmt.Sprintf("✅ %s\n%s: %s",
			tb.translations[lang]["subscriptionActive"],
			tb.translations[lang]["expiresAt"],
			expiresAt)
	} else {
		statusText = tb.translations[lang]["noActiveSubscription"]
	}

	send.SendReplyMessage(tb.Bot, message, &statusText)
	return nil
}
