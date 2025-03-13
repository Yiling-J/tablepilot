package util

import (
	"encoding/json"
	"testing"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/stretchr/testify/require"
)

func TestUtil_ConvertToType(t *testing.T) {
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
		result, err := ConvertToType(test.input, test.tp)
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
