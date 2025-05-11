//go:build !noui
// +build !noui

package ui

import "embed"

//go:embed dist/*
var Dist embed.FS
