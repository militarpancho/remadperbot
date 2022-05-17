package db

import (
	"database/sql"

	"remadperbot/pkg/models"
)

func (db Database) GetAllItemUpdates() (*models.ItemUpdateList, error) {
	list := &models.ItemUpdateList{}
	rows, err := db.Conn.Query("SELECT * FROM item_updates ORDER BY ID DESC")
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var itemUpdate models.ItemUpdate
		err := rows.Scan(&itemUpdate.ID, &itemUpdate.Status, &itemUpdate.CreatedAt)
		if err != nil {
			return list, err
		}
		list.ItemUpdates = append(list.ItemUpdates, itemUpdate)
	}
	return list, nil
}
func (db Database) AddItemUpdate(itemUpdate *models.ItemUpdate) error {
	var createdAt string
	query := `INSERT INTO item_updates (id, status) VALUES ($1, $2) RETURNING created_at`
	err := db.Conn.QueryRow(query, itemUpdate.ID, itemUpdate.Status).Scan(&createdAt)
	if err != nil {
		return err
	}
	itemUpdate.CreatedAt = createdAt
	return nil
}
func (db Database) GetItemUpdateById(itemUpdateId string) (models.ItemUpdate, error) {
	itemUpdate := models.ItemUpdate{}
	query := `SELECT * FROM item_updates WHERE id = $1;`
	row := db.Conn.QueryRow(query, itemUpdateId)
	switch err := row.Scan(&itemUpdate.ID, &itemUpdate.Status, &itemUpdate.CreatedAt); err {
	case sql.ErrNoRows:
		return itemUpdate, ErrNoMatch
	default:
		return itemUpdate, err
	}
}
func (db Database) DeleteItemUpdate(ItemUpdateId string) error {
	query := `DELETE FROM item_updates WHERE id = $1;`
	_, err := db.Conn.Exec(query, ItemUpdateId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}
func (db Database) UpdateItemUpdate(ItemUpdateId string, itemUpdateData models.ItemUpdate) (models.ItemUpdate, error) {
	itemUpdate := models.ItemUpdate{}
	query := `UPDATE item_updates SET id=$1, status=$2 WHERE id=$3 RETURNING id, status, created_at;`
	err := db.Conn.QueryRow(query, itemUpdateData.ID, itemUpdateData.Status, ItemUpdateId).Scan(&itemUpdate.ID, &itemUpdate.Status, &itemUpdate.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return itemUpdate, ErrNoMatch
		}
		return itemUpdate, err
	}
	return itemUpdate, nil
}
