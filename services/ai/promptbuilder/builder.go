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
