package model

import "github.com/google/uuid"

//IsValidUUID checks if UUID is valid
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	if err != nil {
		return false
	}
	return true
}
