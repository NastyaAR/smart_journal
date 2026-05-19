package models

type Achievement struct {
	ID                   int    `json:"id"`
	StudentID            int    `json:"student_id"`
	Title                string `json:"title"`
	Description          string `json:"description"`
	Status               string `json:"status"`
	Confirmed            bool   `json:"confirmed"`
	ConfirmedByTeacherID int    `json:"confirmed_by_teacher_id,omitempty"`
}
