package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowUtils_SanitizeString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty", "", ""},
		{"All valid ASCII", "MyVariable_123", "MyVariable_123"},
		{"Starts with number", "123Identifier", "x4923Identifier"},
		{"Starts with underscore", "_Identifier", "x95Identifier"},
		{"Starts with symbol", "!Identifier", "x33Identifier"},
		{"Contains space", "Hello World", "Hello_World"},
		{"Contains dash", "Hello-World", "Hello_World"},
		{"Contains dot", "Hello.World", "Hello_World"},
		{"Contains symbols", "Test-1@Value#", "Test_1x64Valuex35"},
		{"All symbols", "!@#", "x33x64x35"},
		{"Unicode letters (now invalid)", "你好_世界1", "x20320x22909_x19990x300281"},
		{"Mixed unicode and symbols", "变量-1", "x21464x37327_1"},
		{"Only numbers", "007", "x4807"},
		{"Only underscores", "___", "x95__"},
		{"Starts with non-ASCII letter", "Äbc", "x196bc"}, // Ä is U+00C4 (196)
		{"Contains non-ASCII letter", "abÄc", "abx196c"},
		{"Single non-ASCII char", "é", "x233"}, // é is U+00E9 (233)
		{"Single ASCII digit", "7", "x55"},     // 7 is U+0037 (55)
		{"Single underscore", "_", "x95"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeString(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
