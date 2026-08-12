package SQL_functions

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var SpreadsheetID = os.Getenv("SPREADSHEET_ID")

func CancelBookingToSheet(dateStr, timeStr, userNick string) error {
	ctx := context.Background()
	googleFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(googleFile))
	if err != nil {
		log.Println("❌ Ошибка инициализации Google Sheets:", err)
		return err
	}
	readRange := "Лист1!A:F"

	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		log.Println("❌ Ошибка чтения Google Таблицы:", err)
		return err
	}

	targetRowIndex := -1

	for i, row := range resp.Values {
		if len(row) < 5 {
			continue
		}
		rowDate := fmt.Sprintf("%v", row[0])
		rowTime := fmt.Sprintf("%v", row[1])
		if len(rowTime) >= 5 {
			rowTime = rowTime[:5]
		}
		rowNick := fmt.Sprintf("%v", row[4])

		if rowDate == dateStr && rowTime == timeStr && rowNick == userNick {
			targetRowIndex = i + 1
			break
		}
	}

	if targetRowIndex == -1 {
		log.Printf("⚠️ Запись для отмены не найдена в Google Таблице (Дата: %s, Время: %s, Ник: %s)", dateStr, timeStr, userNick)
		return fmt.Errorf("запись не найдена")
	}
	updateRange := fmt.Sprintf("Лист1!F%d", targetRowIndex)

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{{"cancel"}},
	}
	_, err = srv.Spreadsheets.Values.Update(SpreadsheetID, updateRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Println("❌ Ошибка обновления статуса в Google Таблице:", err)
		return err
	}
	log.Printf("✅ Запись на строке %d успешно переведена в статус 'cancelled'!", targetRowIndex)
	return nil
}

func AppendBookingToSheet(dateStr, timeStr, serviceTitle, userName, status string) error {
	ctx := context.Background()

	googleFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(googleFile))
	if err != nil {
		log.Println("Ошибка инициализации Google Sheets: ", err)
		return err
	}
	rangeName := "Лист1!A:F"
	row := []interface{}{
		dateStr,
		timeStr,
		serviceTitle,
		userName,
		status,
	}
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}
	_, err = srv.Spreadsheets.Values.Append(SpreadsheetID, rangeName, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()

	if err != nil {
		log.Println("Ошибка отправки данных в Google Sheets: ", err)
		return err
	}
	fmt.Println("✅ Запись успешно добавлена в Google Таблицу!")
	return nil
}
