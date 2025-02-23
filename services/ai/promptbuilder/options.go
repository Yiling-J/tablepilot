package promptbuilder

type ColumnOptionsBuilder struct {
	Builder
}

func NewColumnOptionsBuilder(name, description, prompt string) *ColumnOptionsBuilder {
	b := &ColumnOptionsBuilder{}
	b.AddText("Please generate some options for this table Column. Don't duplicate.")
	if prompt != "" {
		el := b.NewXML("Requirement")
		el.CreateText(prompt)
	}
	el := b.NewXML("Column")
	el.CreateAttr("name", name)
	el.CreateAttr("description", description)
	return b
}

func (b *ColumnOptionsBuilder) AddExampleOptions(options []string) {
	if len(options) == 0 {
		return
	}
	el := b.NewXML("ExampleOptions")
	for _, option := range options {
		oel := el.CreateElement("Option")
		oel.CreateAttr("value", option)
	}
}
