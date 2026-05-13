package valueobjects

import (
	"regexp"
	"strings"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || !emailRegex.MatchString(normalized) {
		return Email{}, domainerrors.ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

func (e Email) String() string {
	return e.value
}
