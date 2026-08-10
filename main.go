package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/joho/godotenv"

	tg "db_train33/TGbot_logic"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func main() {
	_ = godotenv.Load()

	cfg := Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Ошибка открытия пула", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	err = db.Ping()
	if err != nil {
		log.Fatal("База данных недоступна", err)
	}

	fmt.Println("Код подключен к базе")

	tg.TGbot_start(db)
}
