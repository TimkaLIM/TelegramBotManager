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
