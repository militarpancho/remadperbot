package models

type ItemUpdate struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
type ItemUpdateList struct {
	ItemUpdates []ItemUpdate `json:"item_updates"`
}
