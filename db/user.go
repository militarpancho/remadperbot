package db

import (
	"database/sql"

	"remadperbot/pkg/models"
)

func (db Database) GetAllUsers() (*models.UserList, error) {
	list := &models.UserList{}
	rows, err := db.Conn.Query("SELECT * FROM users ORDER BY ID DESC")
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
func (db Database) AddUser(User *models.User) error {
	var createdAt string
	query := `INSERT INTO users (ID) VALUES ($1) RETURNING created_at`
	err := db.Conn.QueryRow(query, User.ID).Scan(&createdAt)
	if err != nil {
		return err
	}
	User.CreatedAt = createdAt
	return nil
}
func (db Database) GetUserById(UserId string) (models.User, error) {
	User := models.User{}
	query := `SELECT * FROM users WHERE id = $1;`
	row := db.Conn.QueryRow(query, UserId)
	switch err := row.Scan(&User.ID, &User.CreatedAt); err {
	case sql.ErrNoRows:
		return User, ErrNoMatch
	default:
		return User, err
	}
}
func (db Database) DeleteUser(UserId string) error {
	query := `DELETE FROM users WHERE id = $1;`
	_, err := db.Conn.Exec(query, UserId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}
func (db Database) UpdateUser(UserId int, UserData models.User) (models.User, error) {
	User := models.User{}
	query := `UPDATE items SET id=$1, status=$2 WHERE id=$3 RETURNING id, created_at;`
	err := db.Conn.QueryRow(query, UserData.ID, UserId).Scan(&User.ID, &User.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User, ErrNoMatch
		}
		return User, err
	}
	return User, nil
}
