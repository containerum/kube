package model

import (
	"encoding/base64"

	"git.containerum.net/ch/kube-client/pkg/model"
	"github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

type UserHeaderDataMap map[string]model.UserHeaderData

//ParseUserHeaderData decodes headers for substitutions
func ParseUserHeaderData(str string) (*UserHeaderDataMap, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.WithError(err).WithField("Value", str).Warn(ErrUnableDecodeUserHeaderData)
		return nil, ErrUnableDecodeUserHeaderData
	}
	var userData []model.UserHeaderData
	err = jsoniter.Unmarshal(data, &userData)
	if err != nil {
		log.WithError(err).WithField("Value", string(data)).Warn(ErrUnableUnmarshalUserHeaderData)
		return nil, ErrUnableUnmarshalUserHeaderData
	}
	result := UserHeaderDataMap{}
	for _, v := range userData {
		result[v.ID] = v
	}
	return &result, nil
}
