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
	expected := `<job name="Images-Generation" />
test
For each row in '<Rows>' and each column in '<MissingColumns>', help me generate the missing images as follows:
- Explain the image you intend to generate.
- Provide a text response indicating the corresponding row ID and column ID in <info row_id="xxx" column_id="xxx" /> format.
- **Generate an image** based on the contextual information along with the column’s 'description' using your image-generation capability.
- After generating the image, describe the final result and what the image depicts.`
	require.Equal(t, expected, prompt)
}
