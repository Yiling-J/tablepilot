package utils

import (
	"encoding/csv"
	"io"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func NewCsvReader(r io.Reader) *csv.Reader {
	textReader := transform.NewReader(r, unicode.BOMOverride(transform.Nop))
	return csv.NewReader(textReader)
}
