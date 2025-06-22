package models

import "time"

type Withdrawal struct {
	Order       string
	UserID      string
	Sum         float64
	ProcessedAt time.Time
}
