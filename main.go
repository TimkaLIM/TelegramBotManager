package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/joho/godotenv"

	tg "db_train33/TGbot_logic"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func RunMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("❌ Ошибка создания драйвера миграций: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver)
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации мигратора: %v", err)
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("✅ База данных в актуальном состоянии (миграции не требуются)")
		} else {
			log.Fatalf("Ошибка выполнения миграции!", err)
		}
	} else {
		fmt.Println("🚀 Миграции базы данных успешно применены!")
	}
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

	RunMigrations(db)

	tg.TGbot_start(db)
}
