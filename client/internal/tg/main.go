package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "BulkaVPN/client/proto"
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

	go checkClientsTimeLeft(bot, client)

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
						tgbotapi.NewInlineKeyboardButtonData("Получить VPN ключ (тестовый период 3 дня)", "get_trial_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Подключиться / Продлить ключ", "connect_or_extend"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Мои ключи", "my_keys"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Сменить локацию", "change_location"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Все о BULKA VPN", "about_vpn"),
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

				if resp.CountryServerShadowsocks == "Вы можете выбрать локацию для тестового периода" {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, resp.CountryServerShadowsocks)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Holland, Amsterdam", "choose_holland_trial"),
							tgbotapi.NewInlineKeyboardButtonData("Germany, Frankfurt", "choose_germany_trial"),
						),
					)
					_, err := bot.Send(msg)
					if err != nil {
						return
					}
				} else if resp.CountryServerShadowsocks == "Тестовый период уже был использован" {
					message := "Ваш Тестовый период уже был использован"
					_, err := bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message))
					if err != nil {
						return
					}
				} else {
					message := "Ваш VPN ключ (Shadowsocks / Outline):\n\n" + resp.ShadowsocksVpnConfig +
						"\nЛокация:\n\n" + resp.CountryServerShadowsocks +
						"\nВаш VPN ключ (VLESS):\n" + resp.VlessVpnConfig +
						"\nЛокация:\n\n" + resp.CountryServerVless +
						"\n\nОсталось времени: " + resp.TimeLeft.AsTime().Sub(time.Now()).String()
					_, err := bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message))
					if err != nil {
						return
					}
				}
				_, err = bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
				if err != nil {
					return
				}
			case "choose_holland_trial", "choose_germany_trial":
				telegramID := update.CallbackQuery.From.ID
				country := "Holland, Amsterdam"
				if update.CallbackQuery.Data == "choose_germany_trial" {
					country = "Germany, Frankfurt"
				}

				req := &pb.CreateTrialClientRequest{
					TelegramId:               int64(telegramID),
					StartButton:              false,
					Trial:                    true,
					CountryServerVless:       country,
					CountryServerShadowsocks: country,
				}
				resp, err := client.CreateTrialClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				message := "Ваш VPN ключ: (Shadowsocks / Outline):\n" + resp.ShadowsocksVpnConfig +
					"\nЛокация: " + resp.CountryServerShadowsocks +
					"Ваш VPN ключ: (VLESS):\n" + resp.VlessVpnConfig +
					"\nЛокация: " + resp.CountryServerVless +
					"\nОсталось времени: " + resp.TimeLeft.AsTime().Sub(time.Now()).String()
				_, err = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message))
				if err != nil {
					return
				}

				mainMenu := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Главное меню:")
				mainMenu.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
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
			case "connect_or_extend":
				telegramID := update.CallbackQuery.From.ID

				req := &pb.GetClientRequest{
					TelegramId: int64(telegramID),
				}

				resp, err := client.GetClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				if resp.Clients.ShadowsocksVpnConfig != "" || resp.Clients.VlessVpnConfig != "" {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Заметили, что у Вас уже есть VPN ключ. \n\nХотите продлить его или подключить новый? \n\n*При генерации нового ключа - текущий ключ удаляется \n\nУзнать текущий VPN ключ, его локацию и оставшееся время пользования можно ппо кнопке 'Мой ключ' в Главном меню")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Продлить текущий ключ", "extend_actual"),
							tgbotapi.NewInlineKeyboardButtonData("Подключить новый ключ", "choose_location"),
							tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
						),
					)
					_, err := bot.Send(msg)
					if err != nil {
						return
					}
				}

				if resp.Clients.ShadowsocksVpnConfig == "" || resp.Clients.VlessVpnConfig == "" {
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
				}
			case "extend_actual":

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите продолжительность подключения")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("1 месяц (30 дней)", "duration_30_days"),
						tgbotapi.NewInlineKeyboardButtonData("3 месяца (90 дней)", "duration_90_days"),
						tgbotapi.NewInlineKeyboardButtonData("12 месяцев (365 дней)", "duration_365_days"),
						tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
					),
				)
				_, err = bot.Send(msg)
				if err != nil {
					return
				}
			case "choose_location":
				telegramID := update.CallbackQuery.From.ID

				req := &pb.DeleteClientRequest{
					TelegramId: int64(telegramID),
				}

				resp, err := client.DeleteClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				if resp.Deleted {
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
				}
			case "choose_location_holland", "choose_location_germany":
				location := "Holland, Amsterdam"
				if update.CallbackQuery.Data == "choose_location_germany" {
					location = "Germany, Frankfurt"
				}

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите продолжительность подключения")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("1 месяц (30 дней)", "duration_30_days:"+location),
						tgbotapi.NewInlineKeyboardButtonData("3 месяца (90 дней)", "duration_90_days:"+location),
						tgbotapi.NewInlineKeyboardButtonData("12 месяцев (365 дней)", "duration_365_days:"+location),
						tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
					),
				)
				_, err := bot.Send(msg)
				if err != nil {
					return
				}

			case "duration_30_days:Holland, Amsterdam", "duration_30_days:Germany, Frankfurt",
				"duration_90_days:Holland, Amsterdam", "duration_90_days:Germany, Frankfurt",
				"duration_365_days:Holland, Amsterdam", "duration_365_days:Germany, Frankfurt",
				"duration_30_days", "duration_90_days", "duration_365_days":

				telegramID := update.CallbackQuery.From.ID
				data := update.CallbackQuery.Data
				location := ""
				days := int64(0)

				switch data {
				case "duration_30_days:Holland, Amsterdam":
					location = "Holland, Amsterdam"
					days = 30
				case "duration_30_days:Germany, Frankfurt":
					location = "Germany, Frankfurt"
					days = 30
				case "duration_90_days:Holland, Amsterdam":
					location = "Holland, Amsterdam"
					days = 90
				case "duration_90_days:Germany, Frankfurt":
					location = "Germany, Frankfurt"
					days = 90
				case "duration_365_days:Holland, Amsterdam":
					location = "Holland, Amsterdam"
					days = 90
				case "duration_365_days:Germany, Frankfurt":
					location = "Germany, Frankfurt"
					days = 90
				case "duration_30_days":
					location = "Holland, Amsterdam"
					days = 30
				case "duration_90_days":
					location = "Holland, Amsterdam"
					days = 90
				case "duration_365_days":
					location = "Holland, Amsterdam"
					days = 365
				default:
					location = ""
					days = 0
				}

				req := &pb.CreateClientRequest{
					TelegramId:               int64(telegramID),
					CountryServerShadowsocks: location,
					CountryServerVless:       location,
					TimeLeft:                 timestamppb.New(time.Now().Add(time.Duration(days) * 24 * time.Hour)),
				}
				resp, err := client.CreateClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}
				messageText := fmt.Sprintf("Ваш VPN ключ (Shadowsoks / Outline): %s\n\n"+
					"Локация: %s\n\n"+
					"Ваш VPN ключ (VLESS): %s\n\n"+
					"Локация: %s\n\n"+
					"Осталось времени: %s",
					resp.ShadowsocksVpnConfig,
					resp.CountryServerShadowsocks,
					resp.VlessVpnConfig,
					resp.CountryServerVless,
					resp.TimeLeft.AsTime().Sub(time.Now()).String())

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

			case "my_keys":
				telegramID := update.CallbackQuery.From.ID

				req := &pb.GetClientRequest{
					TelegramId: int64(telegramID),
				}

				resp, err := client.GetClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				messageText := fmt.Sprintf(
					"Ваш VPN ключ (Shadowsoks / Outline): %s\n\n"+
						"Локация: %s\n\n"+
						"Ваш VPN ключ (VLESS): %s\n\n"+
						"Локация: %s\n\n"+
						"Осталось времени: %s",
					resp.Clients.ShadowsocksVpnConfig,
					resp.Clients.CountryServerShadowsocks,
					resp.Clients.VlessVpnConfig,
					resp.Clients.CountryServerVless,
					resp.Clients.TimeLeft.AsTime().Sub(time.Now()).String())

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

			case "change_location":
				telegramID := update.CallbackQuery.From.ID

				req := &pb.GetClientRequest{
					TelegramId: int64(telegramID),
				}

				resp, err := client.GetClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				if resp.Clients.ShadowsocksVpnConfig == "" || resp.Clients.CountryServerVless == "" {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "У Вас нет активного VPN ключа для смены локации")
					_, err = bot.Send(msg)
					if err != nil {
						return
					}
				}

				switch resp.Clients.CountryServerShadowsocks {
				case "Holland, Amsterdam":

					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите локацию:")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Germany, Frankfurt", "update_shadowsocks_location_germany"),
							tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
						),
					)
					_, err = bot.Send(msg)
					if err != nil {
						return
					}

				case "Germany, Frankfurt":

					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите локацию, ")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Holland, Amsterdam", "update_shadowsocks_location_holland"),
							tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
						),
					)
					_, err = bot.Send(msg)
					if err != nil {
						return
					}
				}

				switch resp.Clients.CountryServerVless {
				case "Holland, Amsterdam":

					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите локацию:")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Germany, Frankfurt", "update_vless_location_germany"),
							tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
						),
					)
					_, err = bot.Send(msg)
					if err != nil {
						return
					}

				case "Germany, Frankfurt":

					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите локацию, ")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Holland, Amsterdam", "update_vless_location_holland"),
							tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
						),
					)
					_, err = bot.Send(msg)
					if err != nil {
						return
					}
				}

			case "update_shadowsocks_location_germany", "update_shadowsocks_location_holland":
				telegramID := update.CallbackQuery.From.ID

				location := "Holland, Amsterdam"
				if update.CallbackQuery.Data == "update_shadowsocks_location_germany" {
					location = "Germany, Frankfurt"
				}

				req := &pb.UpdateClientRequest{
					CountryServerShadowsocks: location,
					TelegramId:               int64(telegramID),
				}

				resp, err := client.UpdateClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				messageText := fmt.Sprintf("Ваш новый VPN ключ (Shadowsocks / Outline): %s\n\nЛокация: %s",
					resp.ShadowsocksVpnConfig, resp.CountryServerShadowsocks)

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

			case "update_vless_location_germany", "update_vless_location_holland":
				telegramID := update.CallbackQuery.From.ID

				location := "Holland, Amsterdam"
				if update.CallbackQuery.Data == "update_vless_location_germany" {
					location = "Germany, Frankfurt"
				}

				req := &pb.UpdateClientRequest{
					CountryServerVless: location,
					TelegramId:         int64(telegramID),
				}

				resp, err := client.UpdateClient(context.Background(), req)
				if err != nil {
					_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ошибка: "+err.Error()))
					if err != nil {
						return
					}
					continue
				}

				messageText := fmt.Sprintf("Ваш новый VPN ключ (VLESS): %s\n\nЛокация: %s",
					resp.VlessVpnConfig, resp.CountryServerVless)

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
						tgbotapi.NewInlineKeyboardButtonData("Получить VPN ключ (тестовый период 3 дня)", "get_trial_key"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Подключиться / Продлить ключ", "connect_or_extend"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Мои ключи", "my_keys"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Сменить локацию", "change_location"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Все о BULKA VPN", "about_vpn"),
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

func checkClientsTimeLeft(bot *tgbotapi.BotAPI, client pb.BulkaVPNServiceClient) {
	sentNotifications := make(map[int64]struct {
		notifiedFourDays bool
		notifiedOneDay   bool
	})

	for {
		resp, err := client.SearchClients(context.Background(), &pb.SearchClientsRequest{})
		if err != nil {
			log.Printf("Error fetching clients: %v", err)
			continue
		}

		for _, clients := range resp.Clients {
			telegramID := clients.TelegramId
			now := time.Now()
			timeLeft := clients.TimeLeft.AsTime().Sub(now)

			if timeLeft.Hours() <= 96 && timeLeft.Hours() > 72 {
				if !sentNotifications[telegramID].notifiedFourDays {
					msg := tgbotapi.NewMessage(telegramID, "У Вас осталось четыре дня пользования VPN ключём, рекомендуем заранее продлить ключ для продолжения работы впн")
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Failed to send message to %d: %v", telegramID, err)
					} else {
						sentNotifications[telegramID] = struct {
							notifiedFourDays bool
							notifiedOneDay   bool
						}{true, sentNotifications[telegramID].notifiedOneDay}
					}
				}
			} else if timeLeft.Hours() <= 24 && timeLeft.Hours() > 0 {
				if !sentNotifications[telegramID].notifiedOneDay {
					msg := tgbotapi.NewMessage(telegramID, "У Вас остался один день пользования VPN ключём, для продолжения работы рекомендуем продлить подключение")
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Failed to send message to %d: %v", telegramID, err)
					} else {
						sentNotifications[telegramID] = struct {
							notifiedFourDays bool
							notifiedOneDay   bool
						}{sentNotifications[telegramID].notifiedFourDays, true}
					}
				}
			} else if timeLeft.Hours() <= 0 && clients.IsTrialActiveNow {
				req := &pb.DeleteClientRequest{
					TelegramId:       telegramID,
					IsTrialActiveNow: clients.IsTrialActiveNow,
				}
				_, err := client.DeleteClient(context.Background(), req)
				if err != nil {
					log.Printf("Failed to delete client %d: %v", telegramID, err)
				} else {
					msg := tgbotapi.NewMessage(telegramID, "Время пользования VPN ключём истекло, для продолжения пользования необходимо подключиться заново")
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Failed to send message to %d: %v", telegramID, err)
					}
					delete(sentNotifications, telegramID)
				}
			}
		}

		time.Sleep(10 * time.Minute)
	}
}
