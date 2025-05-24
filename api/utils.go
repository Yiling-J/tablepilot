package api

import (
	"encoding/base64"
	"errors"
	"strings"
)

func DecodeDataURL(dataURL string) ([]byte, error) {
	if !strings.HasPrefix(dataURL, "data:") {
		return nil, errors.New("invalid data URL")
	}

	// Split header and data
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid data URL format")
	}

	metadata, data := parts[0], parts[1]

	isBase64 := strings.HasSuffix(metadata, ";base64")
	if isBase64 {
		return base64.StdEncoding.DecodeString(data)
	} else {
		return nil, errors.New("dataURL is not base64 encoded")
	}
}
