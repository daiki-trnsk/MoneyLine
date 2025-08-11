package infra

import (
    "os"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/daiki-trnsk/MoneyLine/models"
)

var DB *gorm.DB

func InitDB() {
    dsn := os.Getenv("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect database:", err)
    }
    if err := models.AutoMigrate(db); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    DB = db
    log.Println("Connected to Supabase DB!")
}

