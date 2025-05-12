//go:build noui
// +build noui

package ui

import (
	"embed"
)

var Dist embed.FS = embed.FS{}
