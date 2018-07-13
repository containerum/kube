package model

import "github.com/google/uuid"

//IsValidUUID checks if UUID is valid
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return !(err != nil)
}
