package db

import (
	"log"
	"remadperbot/pkg/models"
)

func (db Database) AddUsersItemUpdate(remad_id string, status string, telegram_id string) error {
	usersItemUpdates := models.UsersItemUpdates{}

	itemUpdate, err := db.GetItemUpdateById(remad_id)
	if err == ErrNoMatch {
		itemUpdate = models.ItemUpdate{ID: remad_id, Status: status}
		err = db.AddItemUpdate(&itemUpdate)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	user, err := db.GetUserById(telegram_id)
	if err == ErrNoMatch {
		user = models.User{ID: telegram_id}
		err = db.AddUser(&user)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var createdAt string
	query := `INSERT INTO users_item_updates (item_update_id, user_id) VALUES ($1, $2) RETURNING created_at`
	err = db.Conn.QueryRow(query, itemUpdate.ID, user.ID).Scan(&createdAt)
	if err != nil {
		return err
	}
	usersItemUpdates.User = user
	usersItemUpdates.ItemUpdate = itemUpdate
	usersItemUpdates.CreatedAt = createdAt
	log.Printf("user %s is now monitoring %s", telegram_id, remad_id)
	return nil
}

func (db Database) GetAllUsersByItemUpdate(itemUpdateId string) (models.UserList, error) {
	list := models.UserList{}
	rows, err := db.Conn.Query(`SELECT user_id, created_at FROM users_item_updates WHERE item_update_id = $1`, itemUpdateId)

	if err != nil {
		return list, err
	}
	for rows.Next() {
		var User models.User
		err := rows.Scan(&User.ID, &User.CreatedAt)
		if err != nil {
			return list, err
		}
		list.Users = append(list.Users, User)
	}
	return list, nil
}
