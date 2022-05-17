package models

type User struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}
type UserList struct {
	Users []User `json:"users"`
}
