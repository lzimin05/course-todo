package dto

type UserDTO struct {
	Login    string `json:"login"`
	Username string `json:"username"`
	Email    string `json:"email"`
}