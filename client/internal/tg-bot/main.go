package main

import (
	"context"
	"log"
	"time"

	"BulkaVPN/client/internal/handler"
	"BulkaVPN/client/internal/repository"
	pb "BulkaVPN/client/proto" // Import the correct path to your proto file
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var vpnClient pb.BulkaVPNServiceClient

func main() {
	// Initialize the Telegram bot
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to connect to Telegram API: %v", err)
	}

	// Connect to the gRPC VPN service
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to VPN service: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close connection: %v", err)
		}
	}(conn)

	vpnClient = pb.NewBulkaVPNServiceClient(conn)
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up the Telegram bot update listener
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	// Create a handler instance
	h := handler.Handler{} // Assumes a NewHandler() constructor that returns a *Handler

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Fetch the user by their Telegram ID
		user, err := h.GetUserByTelegramID(context.Background(), int64(update.Message.From.ID))
		if err != nil {
			log.Printf("Error fetching user: %v", err)
			continue
		}

		if user == nil {
			// If user does not exist, create a new one
			user = &repository.User{
				TelegramID: int64(update.Message.From.ID),
				HasTrial:   false,
				TrialEnd:   time.Now(),
			}
			if err := h.SaveUser(context.Background(), user); err != nil {
				log.Printf("Error saving new user: %v", err)
			}
		}

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the VPN Bot! Choose an option:")
			msg.ReplyMarkup = mainMenuKeyboard(user)
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}

		case "Создать ключ":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose a server location:")
			msg.ReplyMarkup = countryKeyboard()
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}

		case "Holland, Amsterdam", "Germany, Frankfurt":
			country := update.Message.Text
			go createVPNKey(bot, update.Message.Chat.ID, country)

		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't understand that command.")
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}

func countryKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Holland, Amsterdam"),
			tgbotapi.NewKeyboardButton("Germany, Frankfurt"),
		),
	)
}

func createVPNKey(bot *tgbotapi.BotAPI, chatID int64, country string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the CreateClient RPC method
	resp, err := vpnClient.CreateClient(ctx, &pb.CreateClientRequest{
		CountryServer: country,
	})
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to create VPN key: "+err.Error())
		_, err := bot.Send(msg)
		if err != nil {
			return
		}
		return
	}

	// Send the ovpn config to the user
	msg := tgbotapi.NewMessage(chatID, "Your VPN key:\n"+resp.OvpnConfig)
	_, err = bot.Send(msg)
	if err != nil {
		return
	}
}

func mainMenuKeyboard(user *repository.User) tgbotapi.ReplyKeyboardMarkup {
	// Row 1: Create key, Remind my key
	row1 := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Создать ключ"),
		tgbotapi.NewKeyboardButton("Напомнить мой ключ"),
	}

	// Row 2: Conditional trial period button
	var row2 []tgbotapi.KeyboardButton
	if !user.HasTrial || user.TrialEnd.Before(time.Now()) {
		row2 = append(row2, tgbotapi.NewKeyboardButton("Пробный период"))
	}

	// Row 3: Delete key, Change country
	row3 := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Удалить ключ"),
		tgbotapi.NewKeyboardButton("Поменять страну"),
	}

	// Combine rows into a keyboard
	return tgbotapi.NewReplyKeyboard(row1, row2, row3)
}

func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update, handler *handler.Handler) {
	ctx := context.Background()

	// Получаем Telegram ID пользователя
	telegramID := update.Message.Chat.ID

	// Ищем пользователя по Telegram ID в базе данных
	user, err := handler.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		log.Println("Error fetching user:", err)
	}

	if user == nil {
		// Если пользователь новый, создаем новую запись
		user = &repository.User{
			TelegramID: telegramID,
			HasTrial:   false,
		}
		err = handler.SaveUser(ctx, user)
		if err != nil {
			log.Println("Error saving new user:", err)
		}
	}

	// Отправляем основное меню в зависимости от статуса пользователя
	msg := tgbotapi.NewMessage(telegramID, "Welcome to the VPN Bot!")
	msg.ReplyMarkup = mainMenuKeyboard(user) // Изменяем клавиатуру в зависимости от статуса пользователя
	_, err = bot.Send(msg)
	if err != nil {
		return
	}
}

func handleTrialPeriod(bot *tgbotapi.BotAPI, update tgbotapi.Update, handler *handler.Handler) {
	ctx := context.Background()

	user, err := handler.GetUserByTelegramID(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println("Error fetching user:", err)
		return
	}

	if user.HasTrial && user.TrialEnd.After(time.Now()) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You have already used the trial period.")
		_, err := bot.Send(msg)
		if err != nil {
			return
		}
		return
	}

	// Give trial VPN key (e.g. call CreateClient with a 3-day period)
	// Assuming CreateClientRequest has some logic for trial periods
	resp, err := vpnClient.CreateClient(ctx, &pb.CreateClientRequest{
		CountryServer: "Holland, Amsterdam",
		Trial:         "true",
	})
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to create trial key: "+err.Error())
		_, err := bot.Send(msg)
		if err != nil {
			return
		}
		return
	}

	// Update user record to reflect trial activation
	user.HasTrial = true
	user.TrialEnd = time.Now().AddDate(0, 0, 3) // Trial for 3 days
	err = handler.SaveUser(ctx, user)
	if err != nil {
		log.Println("Error saving user:", err)
	}

	// Send the VPN config
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your trial VPN key:\n"+resp.OvpnConfig)
	_, err = bot.Send(msg)
	if err != nil {
		return
	}
}
