package system

import "github.com/google/uuid"

type IDGenerator struct{}

func (IDGenerator) NewID() string {
	return uuid.NewString()
}
