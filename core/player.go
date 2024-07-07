package core

import "github.com/google/uuid"

type Player struct {
	UUID     uuid.UUID
	Username string
}
