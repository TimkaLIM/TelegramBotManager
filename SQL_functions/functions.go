package SQL_functions

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Select_dates(db *sql.DB, trueDate time.Time, userID int64) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	var duration_service int

	var service_id int

	err := db.QueryRow("SELECT service_id FROM chern WHERE user_id = $1", userID).Scan(&service_id)
	if err != nil {
		log.Println("Ошибка чтения id сервиса", err)
	}

	err = db.QueryRow("SELECT duration_minutes FROM services WHERE id = $1", service_id).Scan(&duration_service)
	if err != nil {
		log.Println("Ошибка чтения минут из сервисов", err)
	}

	var booked_time []string
	books, err := db.Query("SELECT TO_CHAR(booking_time, 'HH24:MI') FROM bookings WHERE date = $1 AND status IN ('active','dop_time')", trueDate.Format("2006-01-02"))
	if err != nil {
		log.Println("Ошибка чтения из bookings забитого времени", err)
		return tgbotapi.NewInlineKeyboardMarkup(rows...)
	}
	defer books.Close()

	for books.Next() {
		var t string
		if err = books.Scan(&t); err == nil {
			booked_time = append(booked_time, t)
		}
	}
	if err = books.Err(); err != nil {
		log.Println("Ошибка итерации по bookings:", err)
		return tgbotapi.NewInlineKeyboardMarkup(rows...)
	}

	weekday := trueDate.Weekday().String()

	var startStr, endStr string
	var start_time, end_time time.Time

	var minutes int

	err = db.QueryRow("SELECT start_time,end_time,slot_duration_min FROM schedules WHERE day_of_week = $1", weekday).Scan(&startStr, &endStr, &minutes)
	if err != nil {
		log.Println("Ошибка чтения из schedules начального времени, и минут", err)
		return tgbotapi.NewInlineKeyboardMarkup(rows...)
	}
	start_time, err = time.Parse("15:04:05", startStr)
	if err != nil {
		log.Println("Ошибка перевода в нормально время start", err)
	}
	end_time, err = time.Parse("15:04:05", endStr)
	if err != nil {
		log.Println("Ошибка перевода в нормально время end", err)
	}

	duration := time.Duration(minutes) * time.Minute

	var CurrentRow []tgbotapi.InlineKeyboardButton

	for Hour := start_time; !Hour.After(end_time.Add(-time.Duration(duration_service) * time.Minute)); Hour = Hour.Add(duration) {

		timeSlot := Hour.Format("15:04")

		var flag bool = false
		for i := 0; i < len(booked_time); i++ {
			if booked_time[i] == timeSlot {
				flag = true
				break
			}
		}
		if flag == true {
			continue
		}

		Btn := tgbotapi.NewInlineKeyboardButtonData(timeSlot, "time:"+timeSlot+"*"+trueDate.Format("2006-01-02")+"*"+strconv.Itoa(duration_service)+"*"+strconv.Itoa(service_id))

		CurrentRow = append(CurrentRow, Btn)

		if len(CurrentRow) == 3 {
			rows = append(rows, CurrentRow)
			CurrentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	if len(CurrentRow) > 0 {
		rows = append(rows, CurrentRow)
	}

	Btn1 := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
	Row1 := tgbotapi.NewInlineKeyboardRow(Btn1)

	Btn2 := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "write")
	Row2 := tgbotapi.NewInlineKeyboardRow(Btn2)

	rows = append(rows, Row1, Row2)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func СancelTime(db *sql.DB, userID int64, date string, cancelTime string) error {
	_, err := db.Exec("UPDATE bookings SET status = 'cancel' WHERE user_id = $1 AND date = $2 AND TO_CHAR(booking_time,'HH24:MI') = $3", userID, date, cancelTime)
	return err
}
