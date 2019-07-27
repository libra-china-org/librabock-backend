package models

import "time"

type BlockModel struct {
	ID             uint      `gorm:"primary_key" json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	ExpirationAt   time.Time `json:"expiration_at"`
	Version        uint64    `json:"version" gorm:"index:version"`
	Source         string    `json:"source" gorm:"index:source"`
	Destination    string    `json:"destination" gorm:"index:destination"`
	Type           string    `json:"type" gorm:"index:type"`
	Amount         uint64    `json:"amount" `
	GasPrice       uint64    `json:"gas_price" `
	MaxGas         uint64    `json:"max_gas" `
	SequenceNumber uint64    `json:"sequence_number" `
	PublicKey      string    `json:"public_key"`
	MD5            string    `json:"-" `
}

type AccountModel struct {
	Address            string `json:"address"`
	Balance            uint64 `json:"Balance"`
	SequenceNumber     uint64 `json:"sequence_number"`
	SentEventCount     uint64 `json:"sent_event_count"`
	ReceivedEventCount uint64 `json:"received_event_count"`
	AuthenticationKey  string `json:"authentication_key"`
}
