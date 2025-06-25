package models

import "time"

type Withdrawal struct {
	Order       string
	UserID      int
	Sum         float64
	ProcessedAt time.Time
}
