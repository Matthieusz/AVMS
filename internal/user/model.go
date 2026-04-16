package user

import "time"

type User struct {
	ID           uint      `json:"id" db:"id"`
	FirstName    string    `json:"firstName" db:"first_name"`
	LastName     string    `json:"lastName" db:"last_name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose hash in JSON
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
