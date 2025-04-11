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
	prompt := `<job name="Images-Generation-For-Table-Row" />` + "\n"
	p, err := b.Prompt()
	if err != nil {
		return "", err
	}
	prompt += p
	prompt += `Now help me generate the missing images for each row. Here's what you should do:
- For every row in '<Rows>' and for each column in '<MissingColumns>', generate an image based on the contextual information along with the column’s 'description' using your text-to-image capability.
- Before generating each image, also provide a text response indicating the corresponding row ID and column ID in <gen row_id="xxx" column_id="xxx" /> format.`
	return prompt, nil
}
