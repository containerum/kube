package model

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"git.containerum.net/ch/kube-client/pkg/model"

	log "github.com/sirupsen/logrus"
)

var (
	ErrUnableEncodeUserHeaderData    = errors.New("Unbale to encode user header data")
	ErrUnableUnmarshalUserHeaderData = errors.New("Unable unmarshal user header data")
)

func ParseUserHeaderData(str string) (*model.UserHeaderData, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.WithError(err).WithField("Value", str).Warn(ErrUnableEncodeUserHeaderData)
		return nil, ErrUnableEncodeUserHeaderData
	}
	var userData model.UserHeaderData
	err = json.Unmarshal(data, &userData)
	if err != nil {
		log.WithError(err).WithField("Value", string(data)).Warn(ErrUnableUnmarshalUserHeaderData)
		return nil, ErrUnableUnmarshalUserHeaderData
	}
	return &userData, nil
}
