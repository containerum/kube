package model

import (
	"errors"
)

var (
	ErrUnableEncodeUserHeaderData    = errors.New("Unbale to encode user header data")
	ErrUnableUnmarshalUserHeaderData = errors.New("Unable unmarshal user header data")
	ErrUnableConvertServiceList      = errors.New("unable decode service list")
	ErrUnableConvertService          = errors.New("unable convert cubernetes service to user representation")
)
