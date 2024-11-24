package entity

type User struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
