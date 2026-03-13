package entity

import "time"

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID           int64
	Email        string
	Name         string
	Phone        string
	PasswordHash string
	GoogleID     string
	AvatarURL    string
	Role         UserRole
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
