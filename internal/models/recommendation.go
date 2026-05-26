package models

import "time"

type AIGrade struct {
	Subject string `json:"subject"`
	Score   int    `json:"score"`
}

type AIRecommendationRequest struct {
	StudentID      string    `json:"student_id"`
	StudentName    string    `json:"student_name"`
	StudentSurname string    `json:"student_surname"`
	Grades         []AIGrade `json:"grades"`
}

type SubjectRecommendation struct {
	Subject        string `json:"subject"`
	Score          int    `json:"score"`
	Recommendation string `json:"recommendation"`
}

type AIRecommendationResponse struct {
	StudentID       string                  `json:"student_id"`
	StudentName     string                  `json:"student_name"`
	StudentSurname  string                  `json:"student_surname"`
	Level           string                  `json:"level"`
	Strengths       []string                `json:"strengths"`
	Weaknesses      []string                `json:"weaknesses"`
	Recommendations []SubjectRecommendation `json:"recommendations"`
	GeneralAdvice   string                  `json:"general_advice"`
}

type StoredRecommendation struct {
	ID        int                       `json:"id"`
	StudentID int                       `json:"student_id"`
	Payload   *AIRecommendationResponse `json:"payload"`
	CreatedAt time.Time                 `json:"created_at"`
}
