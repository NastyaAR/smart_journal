package models

type Grade struct {
	ID        int `json:"id"`
	StudentID int `json:"student_id"`
	SubjectID int `json:"subject_id"`
	Value     int `json:"value"`
}

type GradeView struct {
	ID          int    `json:"id"`
	StudentID   int    `json:"student_id"`
	StudentName string `json:"student_name"`
	SubjectID   int    `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	Value       int    `json:"value"`
}
