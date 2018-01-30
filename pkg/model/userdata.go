package model

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"
)

type UserHeaderData struct {
	ID     string `json:"id"`     // hosting-internal name
	Label  string `json:"label"`  // user-visible label for the object
	Access string `json:"access"` // one of: "owner", "read", "write", "read-delete", "none"
}

var (
	ErrUnableEncodeUserHeaderData    = errors.New("Unbale to encode user header data")
	ErrUnableUnmarshalUserHeaderData = errors.New("Unable unmarshal user header data")
)

func ParseUserHeaderData(str string) (*UserHeaderData, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.WithError(err).WithField("Value", str).Warn(ErrUnableEncodeUserHeaderData)
		return nil, ErrUnableEncodeUserHeaderData
	}
	var userData UserHeaderData
	err = json.Unmarshal(data, &userData)
	if err != nil {
		log.WithError(err).WithField("Value", string(data)).Warn(ErrUnableUnmarshalUserHeaderData)
		return nil, ErrUnableUnmarshalUserHeaderData
	}
	return &userData, nil
}
