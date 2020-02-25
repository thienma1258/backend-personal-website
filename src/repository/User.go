package repository

import (
	"database/sql"
	"dongpham/src/model"
	_ "github.com/go-sql-driver/mysql"
)

type UserRepository struct {
	db *sql.DB
}

func (up *UserRepository) GetAllUser() []model.User {
	//up.db.Exec()
	return nil
}

func NewUserRepository(db *sql.DB) (*UserRepository, error) {
	return &UserRepository{db: db}, nil
}
