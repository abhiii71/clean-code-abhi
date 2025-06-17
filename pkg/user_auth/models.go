package userauth

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Age       int       `json:"age"`
	DOB       time.Time `json:"-"`
	Gender    string    `json:"gender"`
	Address   Address   `json:"address"`
	Vehicle   Vehicle   `json:"vehicle"`
}
