package models

type UsersItemUpdates struct {
	User       User       `json:"user"`
	ItemUpdate ItemUpdate `json:"item_update"`
	CreatedAt  string     `json:"created_at"`
}
type UsersItemUpdatesList struct {
	UsersItemUpdates []UsersItemUpdates `json:"users_item_updates"`
}
