package tg

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
	"youtube_downloader/pkg/bot/tg/send"
)

const (
	commandStart = "start"
	commandHelp  = "help"
	commandPay   = "pay"

	payMonth    = "pay_month"
	payYear     = "pay_year"
	payLifetime = "pay_lifetime"

	startMessage = "🤖 I'm working! 🤖\n\n" +
		"Hello! I can download video from YouTube, just send a link and choose format\n\n" +
		"📢 Notice! I can download files up to 2Gb\n\n" +
		"📅 The monthly download limit is 5 GB\n\n" +
		"If you want to download more for free, you can sign up for a paid subscription: just enter /pay"

	helpMessage = "I can do the following things:\n\n" +
		"🎬 Download videos from YouTube\n" +
		"🎧 Download audio from YouTube\n" +
		"Just send me a link to the video or audio you want to download."
	defaultMessage = "🤔 I don't know this command. 🤔"
)

// handleCommand handles supported commands
func (tb *TgBot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case commandStart:
		tb.handleStartCommand(message)
	case commandHelp:
		tb.handleHelpCommand(message)
	case commandPay:
		tb.handlePayCommand(message)
	default:
		tb.handleDefaultCommand(message)
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
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Monthly - 100 Rub", payMonth),
		tgbotapi.NewInlineKeyboardButtonData("Yearly - 1000", payYear),
		tgbotapi.NewInlineKeyboardButtonData("Lifetime - 2000", payLifetime),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Choose your subscription plan:")
	msg.ReplyMarkup = keyboard

	if _, err := tb.Bot.Send(msg); err != nil {
		log.Println("Error sending payment options:", err)
	}
}

// processPayment processes the payment based on the selected subscription type
func (tb *TgBot) processPayment(message *tgbotapi.Message, subscriptionType string) {
	subscriptions := map[string]struct {
		Title       string
		Description string
		Amount      int
	}{
		"month": {
			Title:       "Monthly Subscription",
			Description: "Access to downloading without traffic restrictions for one month",
			Amount:      10000, // $2.00
		},
		"year": {
			Title:       "Yearly Subscription",
			Description: "Access to downloading without traffic restrictions for one year",
			Amount:      100000, // $20.00
		},
		"lifetime": {
			Title:       "Lifetime Subscription",
			Description: "Lifetime access to downloading without traffic restrictions",
			Amount:      200000, // $100.00
		},
	}

	subscription, exists := subscriptions[subscriptionType]
	if !exists {
		send.SendMessage(tb.Bot, message, "Invalid subscription type. Please choose 'month', 'year', or 'lifetime'.")
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
	if err != nil {
		log.Printf("Cna't get user: %s, error: %s ", message.From.UserName, err.Error())
	}

	payload := message.SuccessfulPayment.InvoicePayload
	user.Subscription.SubscriptionStatus = "active"
	now := time.Now()
	switch payload {
	case "month":
		user.Subscription.Duration = "1 month"
		user.Subscription.EndSubscription = now.AddDate(0, 1, 0)
	case "year":
		user.Subscription.Duration = "1 year"
		user.Subscription.EndSubscription = now.AddDate(1, 0, 0)
	case "lifetime":
		user.Subscription.Duration = "Lifetime"
		user.Subscription.EndSubscription = time.Time{} // Represents "no end date"
	}

	err = tb.Client.UpdateSubscription(ctx, user)
	if err != nil {
		log.Printf("Error updating user subscription: %s", err.Error())
		send.SendMessage(tb.Bot, message, "Error updating your subscription. Please contact support.")
		return
	}

	send.SendMessage(tb.Bot, message, "Thank you for your payment! Your access has been granted.")
}

// handleStartCommand sends a message with startMessage text
func (tb *TgBot) handleStartCommand(message *tgbotapi.Message) error {
	return send.SendMessage(tb.Bot, message, startMessage)
}

// handleStartCommand sends a message with helpMessage text
func (tb *TgBot) handleHelpCommand(message *tgbotapi.Message) error {
	return send.SendMessage(tb.Bot, message, helpMessage)
}

// handleStartCommand sends a message with defaultMessage text
func (tb *TgBot) handleDefaultCommand(message *tgbotapi.Message) error {
	return send.SendMessage(tb.Bot, message, defaultMessage)
}
