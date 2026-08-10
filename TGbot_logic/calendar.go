package TGbot_logic

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GenerateCalendar(year int, month time.Month) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	monthNames := map[time.Month]string{
		time.January: "Январь", time.February: "Февраль", time.March: "Март",
		time.April: "Апрель", time.May: "Май", time.June: "Июнь",
		time.July: "Июль", time.August: "Август", time.September: "Сентябрь",
		time.October: "Октябрь", time.November: "Ноябрь", time.December: "Декабрь",
	}
	currentDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	prevDate := currentDate.AddDate(0, -1, 0) //-month
	nextDate := currentDate.AddDate(0, 1, 0)  //+month

	prevMonthBtn := tgbotapi.NewInlineKeyboardButtonData("⬅️️", fmt.Sprintf("cal_prev:%s", prevDate.Format("2006-01-02")))
	nextMonthBtn := tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("cal_prev:%s", nextDate.Format("2006-01-02")))
	headerText := fmt.Sprintf("%s %d", monthNames[month], year)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(prevMonthBtn, tgbotapi.NewInlineKeyboardButtonData(headerText, "ignore"), nextMonthBtn))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Пн", "ignore"),
		tgbotapi.NewInlineKeyboardButtonData("Вт", "ignore"),
		tgbotapi.NewInlineKeyboardButtonData("Ср", "ignore"),
		tgbotapi.NewInlineKeyboardButtonData("Чт", "ignore"),
		tgbotapi.NewInlineKeyboardButtonData("Пт", "ignore"),
		tgbotapi.NewInlineKeyboardButtonData("Сб", "ignore"),
		tgbotapi.NewInlineKeyboardButtonData("Вс", "ignore"),
	))

	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	startWeekday := int(firstDay.Weekday())
	if startWeekday == 0 {
		startWeekday = 7
	}

	var currentRow []tgbotapi.InlineKeyboardButton

	for i := 1; i < startWeekday; i++ {
		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(" ", "ignore"))
	}
	for day := 1; day <= lastDay.Day(); day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", day), "date:"+dateStr)
		currentRow = append(currentRow, btn)
		if len(currentRow) == 7 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(currentRow) > 0 {
		for len(currentRow) < 7 {
			currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(" ", "ignore"))
		}
		rows = append(rows, currentRow)
	}

	Btn1 := tgbotapi.NewInlineKeyboardButtonData("🏡 Вернуться в главное меню", "start")
	Row1 := tgbotapi.NewInlineKeyboardRow(Btn1)

	rows = append(rows, Row1)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
