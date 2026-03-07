package entities

import (
	"time"

	"github.com/google/uuid"
)

type Paud struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Subdomain string    `json:"subdomain"`
	Logo      string    `json:"logo,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Context is adalah awokwok informasi tenant untuk request
type Context struct {
	TenantID  uuid.UUID `json:"tenant_id"`
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

// ResolverConfig = konfigurasi resolver
type ResolverConfig struct {
	Domain string
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

type TenantError struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Code     string    `json:"code"`
	Message  string    `json:"message"`
}

func (e TenantError) Error() string {
	return e.Message
}
