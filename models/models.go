package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 一旦最小構成のみ
type Transaction struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreditorID string    `gorm:"index"` // 債権者 LINE ID
	GroupID    string    `gorm:"index"` // グループ LINE ID
	Amount     int64
	Note       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type TransactionDebtor struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TransactionID   uuid.UUID  `gorm:"index"`
	DebtorID        string     `gorm:"index"`
}

// DB接続例（main.goで利用する想定）
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Transaction{}, &TransactionDebtor{})
}
