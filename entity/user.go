package entity

type User struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserToken struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}
