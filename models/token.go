package models

type Token struct {
	Token    string   `json:"token"`
	Username Username `json:"username"`
}
