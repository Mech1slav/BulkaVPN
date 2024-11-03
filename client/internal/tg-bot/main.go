package main

import (
	"context"
	"log"
	"time"

	pb "BulkaVPN/client/proto"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	client := pb.NewBulkaVPNServiceClient(conn)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to get updates: %v", err)
	}

	for update := range updates {
		if update.Message != nil {
			switch update.Message.Text {
			case "/start":
				telegramID := update.Message.From.ID
				req := &pb.CreateTrialClientRequest{
					TelegramId:  int64(telegramID),
					StartButton: true,
				}
				_, err := client.CreateTrialClient(context.Background(), req)
				if err != nil {
					_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка создания клиента: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Главное меню:")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Получить тестовый ключ (пробный период 3 дня)", "get_trial_key"),
					),
				)
				_, err = bot.Send(msg)
				if err != nil {
					return
				}

			default:
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
				if err != nil {
					return
				}
			}
		} else if update.CallbackQuery != nil {
			// Handle callback query for inline button clicks
			switch update.CallbackQuery.Data {
			case "get_trial_key":
				telegramID := update.CallbackQuery.From.ID
				req := &pb.CreateTrialClientRequest{
					TelegramId:  int64(telegramID),
					StartButton: false,
				}
				resp, err := client.CreateTrialClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				if resp.CountryServer == "Вы можете выбрать локацию для пробного периода" {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, resp.CountryServer)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Holland, Amsterdam", "choose_holland"),
							tgbotapi.NewInlineKeyboardButtonData("Germany, Frankfurt", "choose_germany"),
						),
					)
					_, err := bot.Send(msg)
					if err != nil {
						return
					}
				} else {
					message := "Ваш текущий пробный ключ: " + resp.OvpnConfig + "\nОсталось времени: " + resp.TimeLeft.AsTime().Sub(time.Now()).String()
					_, err := bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message))
					if err != nil {
						return
					}
				}
				_, err = bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
				if err != nil {
					return
				}
			case "choose_holland", "choose_germany":
				telegramID := update.CallbackQuery.From.ID
				country := "Holland, Amsterdam"
				if update.CallbackQuery.Data == "choose_germany" {
					country = "Germany, Frankfurt"
				}

				req := &pb.CreateTrialClientRequest{
					TelegramId:    int64(telegramID),
					StartButton:   false,
					Trial:         true,
					CountryServer: country,
				}
				resp, err := client.CreateTrialClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				message := "Ваш пробный VPN ключ: " + resp.OvpnConfig + "\nЛокация: " + resp.CountryServer + "\nОсталось времени: 3 дня"
				_, err = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message))
				if err != nil {
					return
				}

				mainMenu := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Главное меню:")
				mainMenu.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Получить тестовый ключ (пробный период 3 дня)", "get_trial_key"),
					),
				)
				_, err = bot.Send(mainMenu)
				if err != nil {
					return
				}
				_, err = bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
				if err != nil {
					return
				}
			}
		}
	}
}
