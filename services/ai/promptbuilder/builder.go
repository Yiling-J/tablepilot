package promptbuilder

import (
	"fmt"

	"github.com/beevik/etree"
)

type Builder struct {
	elements []any
}

func (b *Builder) AddText(t string) {
	b.elements = append(b.elements, t)
}
func (b *Builder) NewXML(tag string) *etree.Element {
	doc := etree.NewDocument()
	e := doc.CreateElement(tag)
	b.elements = append(b.elements, doc)
	return e
}

func (b *Builder) Prompt() (string, error) {
	prompt := ""
	for _, el := range b.elements {
		switch v := el.(type) {
		case string:
			prompt += fmt.Sprintf("%s\n", v)
		case *etree.Document:
			v.Indent(2)
			s, err := v.WriteToString()
			if err != nil {
				return "", err
			}
			prompt += s
		}
	}
	return prompt, nil
}

func (b *Builder) ImageGenPrompt() (string, error) {
	prompt := `<job name="Images-Generation" />` + "\n"
	p, err := b.Prompt()
	if err != nil {
		return "", err
	}
	prompt += p
	prompt += `For each row in '<Rows>' and each column in '<MissingColumns>', help me generate the missing images as follows:
- Explain the image you intend to generate.
- Provide a text response indicating the corresponding row ID and column ID in <info row_id="xxx" column_id="xxx" /> format.
- **Generate an image** based on the contextual information along with the column’s 'description' using your image-generation capability.
- After generating the image, describe the final result and what the image depicts.`
	return prompt, nil
}
