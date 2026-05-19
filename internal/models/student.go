package models

type Student struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	GroupID  int    `json:"group_id"`
	Tokens   int    `json:"tokens"`
	UserID   int    `json:"user_id"`
}