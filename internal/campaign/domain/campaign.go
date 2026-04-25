package domain

import "time"

type Status string

const (
	StatusActive Status = "active"
	StatusPaused Status = "paused"
)

type Campaign struct {
	ID            string
	AdvertiserID  string
	Name          string
	BudgetCents   int64
	BidPriceCents int64
	Geo           string
	Device        string
	Category      string
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
