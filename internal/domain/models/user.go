package models

type User struct {
	ID       string
	Name     string
	Email    string
	PassHash []byte
	Verified bool
	Avatar   *string
}
