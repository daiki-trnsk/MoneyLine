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

	// 保留→確定フロー用の追加カラム（新規テーブルは作らない）
	RequestID   string `gorm:"index;unique;type:text"` // postback の一意ID (txn/request id)
	DebtorIDs   string `gorm:"type:text"`              // 確定時までの債務者候補をカンマ区切りで保持
	ConfirmedBy string `gorm:"index;type:text"`        // 確定したユーザーの userId

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
