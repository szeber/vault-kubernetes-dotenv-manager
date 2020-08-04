package decoder

import (
	"encoding/base64"
	"errors"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
)

type Decoder struct {
	decodingMethods []string
}

func New(definition config.SecretDefinition) (*Decoder, error) {
	for _, method := range definition.Decoders {
		if !helper.StringInSlice(constants.ValidDecoders[:], method) {
			return nil, errors.New("Invalid decoder method: " + method)
		}
	}

	return &Decoder{decodingMethods: definition.Decoders}, nil
}

func (d *Decoder) DecodeString(s string) ([]byte, error) {
	b := []byte(s)
	var err error

	for _, method := range d.decodingMethods {
		b, err = decode(b, method)
		if nil != err {
			return nil, err
		}
	}

	return b, nil
}

func decode(b []byte, method string) ([]byte, error) {
	switch method {
	case constants.DecoderBase64:
		return decodeBase64(b)
	default:
		return nil, errors.New("Invalid decoder method: " + method)
	}
}

func decodeBase64(s []byte) ([]byte, error) {
	d, err := base64.StdEncoding.DecodeString(string(s))

	if nil != err {
		return nil, err
	}

	return d, nil
}
