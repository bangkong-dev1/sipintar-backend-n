package entities

import (
	"time"

	"github.com/google/uuid"
)

type Paud struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Subdomain string    `db:"subdomain" json:"subdomain"`
	Logo      *string   `db:"logo" json:"logo,omitempty"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Context is adalah awokwok informasi paud untuk request
type Context struct {
	PaudID    uuid.UUID `json:"paud_id"`
	Subdomain string    `json:"subdomain"`
	Status    string    `json:"status"`
}

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusPending   = "pending"
	StatusCancelled = "cancelled"
)

func ValidateStatus(status string) bool {
	switch status {
	case StatusActive, StatusSuspended, StatusPending, StatusCancelled:
		return true
	default:
		return false
	}
}
