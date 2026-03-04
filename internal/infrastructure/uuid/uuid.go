package uuid

import (
	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func Generate() UUID {
	id := uuid.New()
	return UUID(id)
}
