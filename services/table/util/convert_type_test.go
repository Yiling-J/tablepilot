package util

import (
	"encoding/json"
	"testing"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/stretchr/testify/require"
)

func TestUtil_ConvertStringToType(t *testing.T) {
	tests := []struct {
		input      string
		tp         tablecolumn.Type
		expected   any
		expectsErr bool
	}{
		{"hello", tablecolumn.TypeString, "hello", false},
		{"42.5", tablecolumn.TypeNumber, 42.5, false},
		{"42", tablecolumn.TypeInteger, 42, false},
		{"true", tablecolumn.TypeBoolean, true, false},
		{"false", tablecolumn.TypeBoolean, false, false},
		{"[1,2,3]", tablecolumn.TypeArray, []any{float64(1), float64(2), float64(3)}, false},
		{"invalid", tablecolumn.TypeNumber, nil, true},
		{"", tablecolumn.TypeString, "", false},
		{"", tablecolumn.TypeNumber, 0.0, false},
		{"", tablecolumn.TypeInteger, 0, false},
		{"", tablecolumn.TypeBoolean, false, false},
		{"", tablecolumn.TypeArray, []any{}, false},
	}

	for _, test := range tests {
		result, err := ConvertStringToType(test.input, test.tp)
		if test.expectsErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			if test.tp == tablecolumn.TypeArray {
				resultJSON, _ := json.Marshal(result)
				expectedJSON, _ := json.Marshal(test.expected)
				require.JSONEq(t, string(expectedJSON), string(resultJSON))
			} else {
				require.Equal(t, test.expected, result)
			}
		}
	}
}

func TestUtil_ConvertAnyToType(t *testing.T) {
	tests := []struct {
		input    any
		tp       tablecolumn.Type
		expected any
	}{
		{123, tablecolumn.TypeString, "123"},
		{"12.34", tablecolumn.TypeNumber, 12.34},
		{"42", tablecolumn.TypeInteger, 42},
		{"true", tablecolumn.TypeBoolean, true},
		{"false", tablecolumn.TypeBoolean, false},
		{`[1, "two", 3.0]`, tablecolumn.TypeArray, []any{1.0, "two", 3.0}},
		{[]any{1, "two"}, tablecolumn.TypeArray, []any{1, "two"}},
		{"not an array", tablecolumn.TypeArray, []any{}},
		{nil, tablecolumn.TypeString, ""},
		{nil, tablecolumn.TypeNumber, 0.0},
		{nil, tablecolumn.TypeInteger, 0},
		{nil, tablecolumn.TypeBoolean, false},
		{nil, tablecolumn.TypeArray, []any{}},
		{"[1,2,3]", tablecolumn.TypeArray, []any{1.0, 2.0, 3.0}},
		{"{\"key\":\"value\"}", tablecolumn.TypeArray, []any{}},
		{true, tablecolumn.TypeString, "true"},
		{false, tablecolumn.TypeString, "false"},
		{123.45, tablecolumn.TypeInteger, 123},
		{123, tablecolumn.TypeNumber, 123.0},
	}

	for _, test := range tests {
		result := ConvertAnyToType(test.input, test.tp)
		if test.tp == tablecolumn.TypeArray {
			resultJSON, _ := json.Marshal(result)
			expectedJSON, _ := json.Marshal(test.expected)
			require.JSONEq(t, string(expectedJSON), string(resultJSON))
		} else {
			require.Equal(t, test.expected, result)
		}
	}
}
