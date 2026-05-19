package models

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Role    string `json:"role"`
}
