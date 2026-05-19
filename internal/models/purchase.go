package models

import "time"

type Purchase struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	MerchID   int       `json:"merch_id"`
	Title     string    `json:"title,omitempty"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
