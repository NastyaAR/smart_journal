package models

import "time"

type TokenOperation struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	StudentName   string    `json:"student_name,omitempty"`
	TeacherID     *int      `json:"teacher_id,omitempty"`
	TeacherName   string    `json:"teacher_name,omitempty"`
	Amount        int       `json:"amount"`
	OperationType string    `json:"operation_type"`
	Reason        string    `json:"reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
