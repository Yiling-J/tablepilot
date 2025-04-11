package promptbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_ImageGen(t *testing.T) {
	builder := &RowsBuilder{}
	builder.AddText("test")

	prompt, err := builder.ImageGenPrompt()
	require.NoError(t, err)
	expected := `<job name="Images-Generation-For-Table-Row" />
test
Now help me generate the missing images for each row. Here's what you should do:
- For every row in '<Rows>' and for each column in '<MissingColumns>', generate an image based on the contextual information along with the column’s 'description' using your text-to-image capability.
- Before generating each image, also provide a text response indicating the corresponding row ID and column ID in <gen row_id="xxx" column_id="xxx" /> format.`
	require.Equal(t, expected, prompt)
}
