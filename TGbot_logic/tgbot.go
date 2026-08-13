package TGbot_logic

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	base "db_train33/SQL_functions"

	"log"
)

type adminData struct {
	token  string
	admins map[int64]bool
}

func SendMainMenu(bot *tgbotapi.BotAPI, chatID int64, StartKeyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, "Здравствуйте👋\nДобро пожаловать\n"+
		"Выберите действие👇")
	msg.ReplyMarkup = StartKeyboard
	bot.Send(msg)
}

func SendAdminMenu(bot *tgbotapi.BotAPI, chatID int64, AdminKeyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, "Здравствуйте👋\n"+
		"Добро пожаловать в панель администратора\nВыберите действие👇")
	msg.ReplyMarkup = AdminKeyboard
	bot.Send(msg)

}

func TGbot_start(db *sql.DB) {

	//BLOCK 1
	_ = godotenv.Load()

	admins01 := os.Getenv("ADMIN_IDS")

	adminIDs := make(map[int64]bool)

	for _, idStr := range strings.Split(admins01, ",") {

		CleanStr := strings.TrimSpace(idStr) //убирает лишние пробелы

		id, err := strconv.ParseInt(CleanStr, 10, 64)
		if err == nil {
			adminIDs[id] = true
		}

	}

	adminInfo := adminData{
		token:  os.Getenv("TOKEN"),
		admins: adminIDs,
	}
	//BLOCK1

	//BLOCK2
	//InlineButtons
	//InlineButtonsStart
	StartBtn1 := tgbotapi.NewInlineKeyboardButtonData("💅 Записаться", "write")
	StartBtn2 := tgbotapi.NewInlineKeyboardButtonData("✍️ Мои записи", "my_writes")
	StartBtn3 := tgbotapi.NewInlineKeyboardButtonData("📄 Услуги", "price_list")
	//InlineButtonsAdmin
	AdminBtn1 := tgbotapi.NewInlineKeyboardButtonData("📅 Записи на сегодня", "admin_today")
	AdminBtn2 := tgbotapi.NewInlineKeyboardButtonURL("✍️️️️️️ Посмотреть весь график", "https://docs.google.com/spreadsheets/d/1S4_4JmcsO0ozoj5KkeoXAum68VM9DOH7ouh_6du4drw/edit?usp=sharing")

	//InlineRows
	//InlineRowsStart
	StartRow1 := tgbotapi.NewInlineKeyboardRow(StartBtn1)
	StartRow2 := tgbotapi.NewInlineKeyboardRow(StartBtn2)
	StartRow3 := tgbotapi.NewInlineKeyboardRow(StartBtn3)
	//InlineRowsAdmin ⚙️ Настройки
	AdminRow1 := tgbotapi.NewInlineKeyboardRow(AdminBtn1)
	AdminRow2 := tgbotapi.NewInlineKeyboardRow(AdminBtn2)

	//InlineMarkups
	StartKeyboard := tgbotapi.NewInlineKeyboardMarkup(StartRow1, StartRow2, StartRow3)
	AdminKeyboard := tgbotapi.NewInlineKeyboardMarkup(AdminRow1, AdminRow2)

	//BLOCK2

	proxyStr := os.Getenv("TELEGRAM_PROXY")
	var bot *tgbotapi.BotAPI

	if proxyStr == "" {
		log.Println("🚀 ЗАПУСК бота: НАПРЯМУЮ (без прокси)")
		var err error
		bot, err = tgbotapi.NewBotAPI(adminInfo.token)
		if err != nil {
			log.Fatal("❌ Ошибка запуска бота без прокси:", err)
		}
	} else {
		fmt.Printf("🌐 Подключение к Telegram API: ЧЕРЕЗ ПРОКСИ (%s)\n", proxyStr)
		ProxyURL, err := url.Parse(proxyStr)
		if err != nil {
			log.Fatal("❌ Ошибка перевода прокси в URL:", err)
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(ProxyURL),
		}
		httpClient := &http.Client{Transport: transport}

		bot, err = tgbotapi.NewBotAPIWithClient(adminInfo.token, tgbotapi.APIEndpoint, httpClient)
		if err != nil {
			log.Fatal("❌ Не удалось запустить бота через прокси:", err)
		}
	}

	fmt.Printf("✅ Бот успешно авторизован! Учетная запись: @%s\n", bot.Self.UserName)

	var err error
	u := tgbotapi.NewUpdate(0)
	Updates := bot.GetUpdatesChan(u)

	for update := range Updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			userID := update.Message.From.ID
			UserName := update.Message.From.FirstName
			//Блок команд
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					_, err := db.Exec("INSERT INTO users (id,name) VALUES($1,$2) ON CONFLICT (id) DO NOTHING", userID, UserName)
					if err != nil {
						log.Println("Ошибка добавления юзера", err)
					}
					msg := tgbotapi.NewMessage(chatID, "Здравствуйте👋\nДобро пожаловать\n"+
						"Выберите действие👇")
					msg.ReplyMarkup = StartKeyboard
					bot.Send(msg)
				case "admin":
					if adminInfo.admins[update.Message.From.ID] {
						msg := tgbotapi.NewMessage(chatID, "Здравствуйте👋\n"+
							"Добро пожаловать в панель администратора\nВыберите действие👇")
						msg.ReplyMarkup = AdminKeyboard
						bot.Send(msg)
					}
				}
			}
		}
		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			callbackData := update.CallbackQuery.Data
			CallbackID := update.CallbackQuery.ID
			userID := update.CallbackQuery.From.ID

			switch {
			case callbackData == "start":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				SendMainMenu(bot, chatID, StartKeyboard)
			case callbackData == "admin":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				SendAdminMenu(bot, userID, AdminKeyboard)
			case callbackData == "write":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				prices, err := db.Query("SELECT id,title,description,price FROM services WHERE is_active")
				if err != nil {
					log.Println("Ошибка вывода услуг", err)
				}
				defer prices.Close()

				var title, description string
				var price float64
				var id int
				var message string = "🔆 Список услуг:\n"

				var rows [][]tgbotapi.InlineKeyboardButton
				var CurrentRow []tgbotapi.InlineKeyboardButton

				for prices.Next() {
					prices.Scan(&id, &title, &description, &price)
					CurrentMessage := fmt.Sprintf("\n🪭 Название: %s\n💬Описание: %s\n💵Цена: %s₽\n", title, description, strconv.FormatFloat(price, 'f', 0, 64))

					btn := tgbotapi.NewInlineKeyboardButtonData(title, "writedata:"+strconv.Itoa(id))
					CurrentRow = append(CurrentRow, btn)
					rows = append(rows, CurrentRow)
					CurrentRow = []tgbotapi.InlineKeyboardButton{}

					message += CurrentMessage
				}

				Btn1 := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
				Row1 := tgbotapi.NewInlineKeyboardRow(Btn1)
				rows = append(rows, Row1)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

				message += "\nВыберите услугу👇\n"

				msg := tgbotapi.NewMessage(chatID, message)
				msg.ReplyMarkup = keyboard

				bot.Send(msg)

			case strings.HasPrefix(callbackData, "writedata:"):
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				idStr := strings.TrimPrefix(callbackData, "writedata:")
				serviceID, _ := strconv.Atoi(idStr)

				_, err = db.Exec("INSERT INTO chern (user_id,service_id) VALUES($1,$2) ON CONFLICT (user_id) DO UPDATE SET service_id = $2", userID, serviceID)
				if err != nil {
					log.Println("Ошибка создания черновика записи", err)
				}
				now := time.Now()
				calendarKeyboard := GenerateCalendar(now.Year(), now.Month())

				msg := tgbotapi.NewMessage(chatID, "📅 Выберите дату")
				msg.ReplyMarkup = calendarKeyboard

				bot.Send(msg)
			case strings.HasPrefix(callbackData, "cal_prev:"), strings.HasPrefix(callbackData, "cal_next:"):
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				dateStr := strings.TrimPrefix(callbackData, "cal_prev:")
				dateStr = strings.TrimPrefix(dateStr, "cal_next:")

				targetDate, _ := time.Parse("2006-01-02", dateStr)
				newCalendar := GenerateCalendar(targetDate.Year(), targetDate.Month())

				editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, update.CallbackQuery.Message.MessageID, newCalendar)

				bot.Send(editMsg)
			case strings.HasPrefix(callbackData, "date:"):
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				dateStr := strings.TrimPrefix(callbackData, "date:")

				trueDate, _ := time.Parse("2006-01-02", dateStr)

				FreeTimes := base.Select_dates(db, trueDate, userID)

				msg := tgbotapi.NewMessage(chatID, dateStr+"\nВыберите время👇")

				msg.ReplyMarkup = FreeTimes

				bot.Send(msg)

			//ТОТ САМЫЙ КЕЙС БРОНИРОВАНИЯ
			case strings.HasPrefix(callbackData, "time:"):
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				rawData := strings.TrimPrefix(callbackData, "time:")

				parts := strings.Split(rawData, "*")
				var booking_time time.Time
				var booking_date time.Time
				StrTime := parts[0]
				booking_time, err = time.Parse("15:04", StrTime)
				if err != nil {
					log.Println("Ошибка перевода времени", err)
				}
				StrDate := parts[1]
				booking_date, err = time.Parse("2006-01-02", StrDate)
				if err != nil {
					log.Println("Ошибка перевода времени", err)
				}
				strDuration := parts[2]
				strServiceID := parts[3]

				ServiceID, err := strconv.Atoi(strServiceID)
				if err != nil {
					log.Println("Ошибка перевода сервисID в число из строки", err)
				}
				var ServiceTitle string

				err = db.QueryRow("SELECT title FROM services WHERE id = $1", ServiceID).Scan(&ServiceTitle)

				duration, err := strconv.Atoi(strDuration)
				if err != nil {
					log.Println("Ошибка перевода длины сеанса из строки в число", err)
				}
				var duration2 int
				err = db.QueryRow("SELECT slot_duration_min FROM schedules WHERE id = 1").Scan(&duration2)
				if err != nil {
					log.Println("ОШИБКА чтения минимального времени", err)
				}
				_, err = db.Exec("INSERT INTO bookings (user_id,booking_time,date,service_id) VALUES($1,$2,$3,$4)", userID, booking_time, booking_date, ServiceID)
				if err != nil {
					log.Println("Ошибка бронирования 1", err)
				}
				///забитие соседнего времени
				duration3 := duration
				currentTime := booking_time
				for duration3 > duration2 {
					currentTime = currentTime.Add(time.Duration(duration2) * time.Minute)

					_, err = db.Exec("INSERT INTO bookings (user_id,booking_time,date,status,service_id) VALUES($1,$2,$3,'dop_time',$4)", userID, currentTime, booking_date, ServiceID)
					if err != nil {
						log.Println("Ошибка бронирования 2", err)
					}
					duration3 -= duration2
				}
				for i := range adminInfo.admins {
					message := fmt.Sprintf("✍️ К вам записался новый клиент!\n📅 Дата: %s\n🕑 Время: %s\n⭐️ %s\n", StrDate[:10], StrTime, ServiceTitle)
					AdminMsg := tgbotapi.NewMessage(i, message)
					bot.Send(AdminMsg)

				}

				_, err = db.Exec("DELETE FROM chern WHERE user_id = $1", userID)
				if err != nil {
					log.Println("Ошибка удаления пользователя из черновика запроса", err)
				}

				msg := tgbotapi.NewMessage(chatID, "✍️ Вы были успешно записаны\n"+"⭐️ "+ServiceTitle+"\n🕑 Время записи:"+StrTime)

				Btn1 := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
				Row1 := tgbotapi.NewInlineKeyboardRow(Btn1)
				Keyboard := tgbotapi.NewInlineKeyboardMarkup(Row1)

				msg.ReplyMarkup = Keyboard

				bot.Send(msg)

				//ДОБАВЛЕНИЕ ЗАПИСИ В ГУГЛ ТАБЛИЦУ
				go func() {

					userNick := update.CallbackQuery.From.UserName
					if userNick != "" {
						userNick = "@" + userNick
					} else {
						userNick = update.CallbackQuery.From.FirstName
					}
					sheetErr := base.AppendBookingToSheet(StrDate, StrTime, ServiceTitle, userNick, "active")
					if sheetErr != nil {
						log.Println("⚠️ [ФОН] Ошибка добавления записи в Google Таблицу:", sheetErr)
					}
				}()
			case callbackData == "my_writes":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				var message string = ""

				var booking_time string
				var booking_date string
				var service_id int

				times, err := db.Query("SELECT booking_time,date,service_id FROM bookings WHERE user_id = $1 AND status = 'active' ORDER BY date ASC, booking_time ASC", userID)
				if err != nil {
					log.Println("Ошибка чтения записей клиента", err)
				}
				defer times.Close()

				var current_date string

				for times.Next() {

					times.Scan(&booking_time, &booking_date, &service_id)
					var serviceTitle string

					err = db.QueryRow("SELECT title FROM services WHERE id = $1", service_id).Scan(&serviceTitle)
					if err != nil {
						log.Println("Ошибка чтения названия записи", err)
					}
					service_id = 0
					if booking_date != current_date {
						message += fmt.Sprintf("\n📅 Дата: %s\n", booking_date[:10])
						current_date = booking_date
					}
					message += fmt.Sprintf("🕑 Время: %s\n", booking_time[:5])
					message += fmt.Sprintf("⭐️ Услуга: %s\n===================\n", serviceTitle)

				}
				Btn1 := tgbotapi.NewInlineKeyboardButtonData("🚫 Отменить запись", "dont_write")
				Btn2 := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")

				Row1 := tgbotapi.NewInlineKeyboardRow(Btn1)
				Row2 := tgbotapi.NewInlineKeyboardRow(Btn2)

				Keyboard := tgbotapi.NewInlineKeyboardMarkup(Row1, Row2)

				msg := tgbotapi.NewMessage(chatID, "Ваши записи👇\n"+message)
				msg.ReplyMarkup = Keyboard

				bot.Send(msg)
			case callbackData == "dont_write":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				msg := tgbotapi.NewMessage(chatID, "📅 Выберите дату записи👇")

				var booking_date string

				dates, err := db.Query("SELECT DISTINCT date FROM bookings WHERE user_id = $1 AND status = 'active'", userID)
				if err != nil {
					log.Println("Ошибка чтения даты записи", err)
				}
				defer dates.Close()

				var rows [][]tgbotapi.InlineKeyboardButton

				var CurrentRow []tgbotapi.InlineKeyboardButton

				for dates.Next() {
					dates.Scan(&booking_date)

					btn := tgbotapi.NewInlineKeyboardButtonData(booking_date[:10], "cancel_date:"+booking_date[:10])

					CurrentRow = append(CurrentRow, btn)

					if len(CurrentRow) == 2 {
						rows = append(rows, CurrentRow)
						CurrentRow = []tgbotapi.InlineKeyboardButton{}
					}
				}
				if len(CurrentRow) > 0 {
					rows = append(rows, CurrentRow)
				}

				btn := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
				row := tgbotapi.NewInlineKeyboardRow(btn)

				rows = append(rows, row)

				Keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
				msg.ReplyMarkup = Keyboard

				bot.Send(msg)

			case strings.HasPrefix(callbackData, "cancel_date:"):
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				date := strings.TrimPrefix(callbackData, "cancel_date:")

				msg := tgbotapi.NewMessage(chatID, "🕑 Выберите время👇")

				times, err := db.Query("SELECT booking_time FROM bookings WHERE user_id = $1 AND date = $2 AND status = 'active'", userID, date)
				if err != nil {
					log.Println("Ошибка чтения времени записи на дату "+date, err)
				}
				defer times.Close()

				var booking_time string

				var rows [][]tgbotapi.InlineKeyboardButton

				var CurrentRow []tgbotapi.InlineKeyboardButton

				for times.Next() {
					times.Scan(&booking_time)

					btn := tgbotapi.NewInlineKeyboardButtonData(booking_time[:5], "cancel_time:"+date+"*"+booking_time[:5])

					CurrentRow = append(CurrentRow, btn)

					if len(CurrentRow) == 2 {
						rows = append(rows, CurrentRow)
						CurrentRow = []tgbotapi.InlineKeyboardButton{}
					}
				}

				if len(CurrentRow) > 0 {
					rows = append(rows, CurrentRow)
				}

				btn := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
				row := tgbotapi.NewInlineKeyboardRow(btn)

				rows = append(rows, row)

				Keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

				msg.ReplyMarkup = Keyboard

				bot.Send(msg)
			case strings.HasPrefix(callbackData, "cancel_time:"):
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				rawData := strings.TrimPrefix(callbackData, "cancel_time:")

				parts := strings.Split(rawData, "*")

				booking_date := parts[0]
				booking_time := parts[1]

				btn := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
				row := tgbotapi.NewInlineKeyboardRow(btn)
				Keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

				msg := tgbotapi.NewMessage(chatID, "Запись отменена❌")
				msg.ReplyMarkup = Keyboard

				err = base.СancelTime(db, userID, booking_date, booking_time)
				if err != nil {
					log.Println("Ошибка отмены записи", err)
				} else {
					bot.Send(msg)

					go func() {
						userNick := update.CallbackQuery.From.UserName
						if userNick != "" {
							userNick = "@" + userNick
						} else {
							userNick = update.CallbackQuery.From.FirstName
						}
						cancelErr := base.CancelBookingToSheet(booking_date, booking_time, userNick)
						if cancelErr != nil {
							log.Println("⚠️ [ФОН] Не удалось отменить запись в Google Таблице:", cancelErr)
						}
					}()
				}

			case callbackData == "price_list":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				prices, err := db.Query("SELECT title,description,price FROM services WHERE is_active")
				if err != nil {
					log.Println("Ошибка вывода услуг", err)
				}
				defer prices.Close()

				var title, description string
				var price float64
				var message string = "🔆 Список услуг:\n"

				for prices.Next() {
					prices.Scan(&title, &description, &price)
					CurrentMessage := fmt.Sprintf("\n🪭 Название: %s\n💬Описание: %s\n💵Цена: %s₽\n", title, description, strconv.FormatFloat(price, 'f', 0, 64))
					message += CurrentMessage
				}
				msg := tgbotapi.NewMessage(chatID, message)

				btn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "start")
				row := tgbotapi.NewInlineKeyboardRow(btn)
				keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

				msg.ReplyMarkup = keyboard

				bot.Send(msg)
			case callbackData == "admin_today":
				bot.Request(tgbotapi.NewCallback(CallbackID, ""))

				now := time.Now()
				todayStr := now.Format("2006-01-02")
				var service_id int
				var todayTime string
				var serviceTitle string
				var message string = "✍️ Ваши записи на сегодня👇\n"

				times, err := db.Query("SELECT booking_time,service_id FROM bookings WHERE date = $1 AND status = 'active'", todayStr)
				if err != nil {
					log.Println("Ошибка чтения времени и сервиса АДМИН", err)
				}
				for times.Next() {
					times.Scan(&todayTime, &service_id)
					err = db.QueryRow("SELECT title FROM services WHERE id = $1", service_id).Scan(&serviceTitle)
					if err != nil {
						log.Println("Ошибка чтения названия услуги в сервисах", err)
					}
					message += fmt.Sprintf("🕑 Время: %s\n⭐️ %s\n\n", todayTime[:5], serviceTitle)
				}
				msg := tgbotapi.NewMessage(chatID, message)

				btn := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "admin")

				row := tgbotapi.NewInlineKeyboardRow(btn)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

				msg.ReplyMarkup = keyboard

				bot.Send(msg)

			}

		}

	}
}
