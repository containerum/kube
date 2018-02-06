package model

import (
	"encoding/base64"
	"encoding/json"

	"git.containerum.net/ch/kube-client/pkg/model"

	log "github.com/sirupsen/logrus"
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
