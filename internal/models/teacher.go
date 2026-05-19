package models

type Teacher struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	UserID int    `json:"user_id"`
}