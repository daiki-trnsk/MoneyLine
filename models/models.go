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
	// Note removed: memo is no longer stored per new requirement
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	Debtors []TransactionDebtor `gorm:"constraint:OnDelete:CASCADE;"`
}

type TransactionDebtor struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TransactionID uuid.UUID `gorm:"index;"`
	DebtorID      string    `gorm:"index"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`

	Transaction Transaction `gorm:"foreignKey:TransactionID;references:ID"`
}

type JoinGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GroupID   string    `gorm:"index;unique"`
	Number    int64
	IsNowIn   bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Transaction{}, &TransactionDebtor{}, &JoinGroup{})
}
