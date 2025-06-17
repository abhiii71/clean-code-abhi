package userauth

import (
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Date time.Time

// UserRegisterRequest represents the data required for user registration
type UserRegisterRequest struct {
	ID        int    `json:"id" example:"101"` // optional
	FirstName string `json:"first_name" example:"Abhishek"`
	LastName  string `json:"last_name" example:"Verma"`
	Email     string `json:"email" example:"abhishek@example.com"`
	Password  string `json:"password" example:"Password@123"`
	DOB       Date   `json:"dob" example:"2000-01-01"`
	Gender    string `json:"gender" example:"male/female"`
}

// for user_information
type UserInformationRequest struct {
	ID        int     `json:"id"`
	Address   Address `json:"address"`
	Vehicle   Vehicle `json:"vehicle"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// for address field inside userinformation
type Address struct {
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
	Country    *string `json:"country"`
}

// for address field inside userinformation
type Vehicle struct {
	Car  *bool `json:"car"`
	Bike *bool `json:"bike"`
}

// enum types
type Gender string

const (
	Male    Gender = "M"
	Female  Gender = "F"
	Shemale Gender = "S"
)

// for login
// these examples are not necessary but we follow
type UserLoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securePassword123"`
}

// emailregex
var emailregex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z.-]+\.[a-zA-Z]{2,}$`)

// function to handle the dob time
func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	// remove quotes from JSON string input
	s = strings.Trim(s, `"`)

	//
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = Date(parsed)
	return nil
}

// marshal
func (d Date) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	return []byte(`"` + t.Format("2006-01-02") + `"`), nil
}

func (u *UserRegisterRequest) Valid() url.Values {
	err := url.Values{}

	// first name
	if len(u.FirstName) == 0 {
		err.Add("first_name", " name cannot be empty")
	} else if len(u.FirstName) < 2 {
		err.Add("first_name", "name too short")
	}

	// email regex
	if len(u.Email) == 0 {
		err.Add("email: ", "email is required")
	} else {
		if !emailregex.MatchString(u.Email) {
			err.Add("email", "invalid email format")
		}
		if regexp.MustCompile(`\.\.`).MatchString(u.Email) {
			err.Add("email", "email cannot contains double dots")
		}
	}

	// password
	if len(u.Password) == 0 {
		err.Add("password", "password is required")
	} else if len(u.Password) < 8 {
		err.Add("password", "must be at least 8 characters")
	} else {
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(u.Password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(u.Password)
		hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(u.Password)
		if !hasUpper || !hasLower || !hasSpecial {
			err.Add("password", "must include 1 uppercase, 1 lowercase, 1 special character")
			// regex for strong password
			// regex := `^(?=*[a-z])(?=.*[A-Z])(?=.*\W{8,}$)`
		}
	}
	// dob
	today := time.Now()
	dob := time.Time(u.DOB)
	age := today.Year() - dob.Year()
	if today.Month() < dob.Month() || today.Month() == dob.Month() && today.Day() < dob.Day() {
		age--
	}
	if age < 18 {
		err.Add("dob", "must be at least 18 years old")
	}

	// gender
	switch Gender(u.Gender) {
	case Male, Female, Shemale:
		// ok
	default:
		err.Add("gender", "must be one of M, F or S")
	}

	return err
}
