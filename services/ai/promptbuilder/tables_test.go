package promptbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_Tables(t *testing.T) {
	builder := NewTablesBuilder("foo")

	prompt, err := builder.Prompt()
	require.NoError(t, err)
	expected := `I am building a system based on **Domain-Driven Design (DDD)**. The first step is to design **domain tables** that represent the core concepts of this system. For each generatad table, please tell me table name, description and depends. Each table should represent a **high-level domain concept**, not low-level technical constructs.

<Example>
{
  "data": [
    {"name": "character", "description": "Characters in the story.", "depends":[]},
    {"name": "planet", "description": "Planets or regions where the story takes place.", "depends":[]}
  ]
}
</Example>

<Rules>
1. The user prompt describing the system is provided in the <UserPrompt> section.
2. Table name **must** start with a letter and contain only letters, numbers, or underscores.
3. **Group related domain info into a single table** whenever possible. For example, if the domain is about recipes, don't split it into "Recipe", "Ingredient", and "RecipeIngredient", just use a single "Recipe" table, and include ingredients as a column.
4. **Do not create intermediate or join tables.** Only include **main domain tables** whose data can be generated independently using AI. For example, avoid tables like "UserLikes" or "UserLogs" unless they are true domain concepts.
5. **Avoid overly trivial or implementation-specific tables.** Focus on meaningful, rich, and generatable content.
6. The "depends" array should list other table names that this table **relies on for meaningful data**. These dependencies must be created **first**.
7. Assume all tables will be filled using **AI-powered content generation**, so each one must be clear, high-level, and self-contained.
</Rules>
<UserPrompt>foo</UserPrompt>
`
	require.Equal(t, expected, prompt)
}

func TestPromptBuilder_TablesPolish(t *testing.T) {
	builder := NewTablesPolishBuilder("tables", "foo")
	prompt, err := builder.Prompt()
	require.NoError(t, err)
	expected := `Here is a JSON data of a list of tables, Each table has name, description and depends. The "depends" array should list other table names that this table **relies on for meaningful data**. These dependencies must be created **first**. Please modify this list based on <Requirement>
<Data>tables</Data>
<Requirement>foo</Requirement>
`
	require.Equal(t, expected, prompt)
}
