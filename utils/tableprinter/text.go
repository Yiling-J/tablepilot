// This code has been modified from its original form by Yiling-J.

// MIT License

// Copyright (c) 2021 GitHub Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

package tableprinter

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/spf13/cast"
)

const (
	ellipsis            = "..."
	minWidthForEllipsis = len(ellipsis) + 2
)

// DisplayWidth calculates what the rendered width of string s will be.
func DisplayWidth(s string) int {
	return lipgloss.Width(s)
}

// Truncate returns a copy of the string s that has been shortened to fit the maximum display width.
func Truncate(maxWidth int, s string) string {
	w := DisplayWidth(s)
	if w <= maxWidth {
		return s
	}
	tail := ""
	if maxWidth >= minWidthForEllipsis {
		tail = ellipsis
	}
	r := truncate.StringWithTail(s, cast.ToUint(maxWidth), tail)
	if DisplayWidth(r) < maxWidth {
		r += " "
	}
	return r
}

// PadRight returns a copy of the string s that has been padded on the right with whitespace to fit
// the maximum display width.
func PadRight(maxWidth int, s string) string {
	if padWidth := maxWidth - DisplayWidth(s); padWidth > 0 {
		s += strings.Repeat(" ", padWidth)
	}
	return s
}
