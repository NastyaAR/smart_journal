package models

import "time"

type Grade struct {
	ID         int       `json:"id"`
	StudentID  int       `json:"student_id"`
	SubjectID  int       `json:"subject_id"`
	Value      int       `json:"value"`
	LessonDate string    `json:"lesson_date,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

type GradeView struct {
	ID          int       `json:"id"`
	StudentID   int       `json:"student_id"`
	StudentName string    `json:"student_name"`
	SubjectID   int       `json:"subject_id"`
	SubjectName string    `json:"subject_name"`
	Value       int       `json:"value"`
	LessonDate  string    `json:"lesson_date,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}
