package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	TRANSACTION_STATUS_PENDING    = 1
	TRANSACTION_STATUS_DONE       = 2
	TRANSACTION_STATUS_ERROR      = 3
	TRANSACTION_STATUS_PENDING_CB = 4
)

const (
	TRANSACTION_KIND_NFT   = 1
	TRANSACTION_KIND_TOKEN = 2
)

type Transaction struct {
	gorm.Model
	Uid           string    `gorm:"column:uid; index" binding:"required"`
	Kind          string    `gorm:"column:kind" binding:"required"`
	Status        byte      `gorm:"column:status" binding:"required"`
	WalletAddr    string    `json:"column:wallet_addr" binding:"required"`
	WalletKind    string    `json:"column:wallet_kind" binding:"required"`
	CallbackUri   string    `json:"column:callback_uri" binding:"required"`
	IpfsCid       string    `json:"column:ipfs_cid"`
	TokenQuantity int       `json:"column:token_quantity"`
	Metadata      JSONMap   `gorm:"column:metadata"`
	Cost          string    `json:"column:cost"`
	ErrorCount    int       `json:"column:error_count"`
	LastErrorAt   time.Time `json:"column:last_error_at"`
	LastError     string    `json:"column:last_error"`
}
