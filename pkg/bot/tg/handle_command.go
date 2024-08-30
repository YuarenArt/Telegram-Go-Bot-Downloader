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
)

// TODO првоекра статуса перед оплатой

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
	lang := "en"
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
		log.Fatalf("Error loading .env file: %v", err)
	}
	providerToken := os.Getenv("PROVIDER_TOKEN")

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
	log.Printf("Successful payment from %s, amount: %d", message.From.UserName, message.SuccessfulPayment.TotalAmount)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	user, err := tb.Client.GetUser(ctx, message.From.UserName)
	if err != nil || user == nil {
		log.Printf("Can't get user: %s, error: %s ", message.From.UserName, err.Error())
	}

	lang := message.From.LanguageCode
	payload := message.SuccessfulPayment.InvoicePayload
	user.Subscription.SubscriptionStatus = "active"

	switch payload {
	case "month":
		user.Subscription.Duration = "month"
		user.Subscription.EndSubscription.AddDate(0, 1, 0)
	case "year":
		user.Subscription.Duration = "year"
		user.Subscription.EndSubscription.AddDate(1, 0, 0)
	case "lifetime":
		user.Subscription.Duration = "lifetime"
		user.Subscription.EndSubscription.AddDate(900, 0, 0) // Represents "no end date"
	}

	err = tb.Client.UpdateSubscription(ctx, user)
	if err != nil {
		log.Printf("Error updating user subscription: %s", err.Error())
		send.SendMessage(tb.Bot, message, tb.translations[lang]["errorUpdatingSubscription"])
		return
	}

	send.SendMessage(tb.Bot, message, tb.translations[lang]["thankYouForPayment"])
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

// UserStatus send user's subscription status and subscription expiration date if active
func (tb *TgBot) UserStatus(message *tgbotapi.Message, lang string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	user, err := tb.Client.GetUser(ctx, message.From.UserName)
	if err != nil || user == nil {
		log.Println("can't get user: " + message.From.UserName)
		return send.SendMessage(tb.Bot, message, tb.translations[lang]["errorFindStatus"])
	}

	statusText := tb.translations[lang]["userStatus"]
	expireText := tb.translations[lang]["expireSubscription"]

	text := fmt.Sprintf("%s %s.", statusText, user.Subscription.SubscriptionStatus)
	if user.Subscription.SubscriptionStatus == "active" {
		text += fmt.Sprintf(" %s %s", expireText, user.Subscription.EndSubscription.Format("2006-01-02"))
	}

	return send.SendMessage(tb.Bot, message, text)
}
