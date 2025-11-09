package dto

type TokenResponse struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	EmailOrLogin string `json:"emailorlogin"`
	Password     string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}