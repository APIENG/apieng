package models

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	Iid       string
	Email     string
	FirstName string
	LastName  string
	Password  string
	Apikey    sql.NullString
}

func (u *Users) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *Users) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
