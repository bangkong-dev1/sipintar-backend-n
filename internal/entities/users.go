package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	PaudID    *uuid.UUID `db:"paud_id" json:"paud_id,omitempty"`
	RoleID    *uuid.UUID `db:"role_id" json:"role_id,omitempty"`
	Name      string     `db:"name" json:"name"`
	Email     string     `db:"email" json:"email"`
	Password  string     `db:"password" json:"-"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// UserWithRole untuk response yang include role info
type UserWithRole struct {
	User
	RoleName *string `db:"role_name" json:"role_name,omitempty"`
}
