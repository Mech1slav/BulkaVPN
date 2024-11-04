package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "BulkaVPN/client/proto"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
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
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Подключиться / Продлить ключ", "connect_or_extend_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Мой ключ", "my_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Сменить локацию", "change_location"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Все о BALT VPN", "about_vpn"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Донаты", "donations"),
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
			case "connect_or_extend_key":
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы можете выбрать локацию для подключения (Локацию всегда можно изменить - это бесплатно)")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Holland, Amsterdam", "choose_location_holland"),
						tgbotapi.NewInlineKeyboardButtonData("Germany, Frankfurt", "choose_location_germany"),
						tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
					),
				)
				_, err := bot.Send(msg)
				if err != nil {
					return
				}
			case "choose_location_holland", "choose_location_germany":
				location := "Holland, Amsterdam"
				if update.CallbackQuery.Data == "choose_location_germany" {
					location = "Germany, Frankfurt"
				}

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите продолжительность подключения")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("30 дней", "duration_30_days:"+location),
						tgbotapi.NewInlineKeyboardButtonData("90 дней", "duration_90_days:"+location),
						tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
					),
				)
				_, err := bot.Send(msg)
				if err != nil {
					return
				}

			case "duration_30_days:Holland, Amsterdam", "duration_30_days:Germany, Frankfurt",
				"duration_90_days:Holland, Amsterdam", "duration_90_days:Germany, Frankfurt":

				telegramID := update.CallbackQuery.From.ID
				data := update.CallbackQuery.Data
				location := ""
				days := int64(0)

				if data == "duration_30_days:Holland, Amsterdam" {
					location = "Holland, Amsterdam"
					days = 30
				} else if data == "duration_30_days:Germany, Frankfurt" {
					location = "Germany, Frankfurt"
					days = 30
				} else if data == "duration_90_days:Holland, Amsterdam" {
					location = "Holland, Amsterdam"
					days = 90
				} else if data == "duration_90_days:Germany, Frankfurt" {
					location = "Germany, Frankfurt"
					days = 90
				}

				req := &pb.CreateClientRequest{
					TelegramId:    int64(telegramID),
					CountryServer: location,
					TimeLeft:      timestamppb.New(time.Now().Add(time.Duration(days) * 24 * time.Hour)),
				}
				resp, err := client.CreateClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}
				messageText := fmt.Sprintf("Ваш VPN ключ: %s\nЛокация: %s\nОсталось времени: %s",
					resp.OvpnConfig, resp.CountryServer, resp.TimeLeft.AsTime().Sub(time.Now()).String())

				mainMenuButton := tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(mainMenuButton),
				)

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, messageText)
				msg.ReplyMarkup = keyboard
				_, err = bot.Send(msg)
				if err != nil {
					return
				}

			case "main_menu":
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Главное меню:")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Получить тестовый ключ (пробный период 3 дня)", "get_trial_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Подключиться / Продлить ключ", "connect_or_extend_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Мой ключ", "my_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Сменить локацию", "change_location"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Все о BALT VPN", "about_vpn"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Донаты", "donations"),
					),
				)
				_, err2 := bot.Send(msg)
				if err2 != nil {
					return
				}
			}
		}
	}
}
