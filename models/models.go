package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 一旦最小構成のみ
type Transaction struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreditorID string    `gorm:"index"`
	GroupID    string    `gorm:"index"`
	Amount     int64
	Note       string
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

type TransactionDebtor struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TransactionID uuid.UUID `gorm:"index;not null;constraint:OnDelete:CASCADE"`
	DebtorID      string    `gorm:"index"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Transaction{}, &TransactionDebtor{})
}
