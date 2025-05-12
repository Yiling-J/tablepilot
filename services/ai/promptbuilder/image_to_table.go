package promptbuilder

import (
	"fmt"
	"strings"
)

type ImageToTableBuilder struct {
	Builder
}

func NewNewImageToTableBuilder(prompt string) *ImageToTableBuilder {
	ib := &ImageToTableBuilder{}
	ib.AddText("Please extract a table from the provided image. All extracted data should be placed in a single table with a consistent schema.")
	if len(prompt) > 0 {
		ib.AddText("## User Requirements:")
		el := ib.NewXML("Requirement")
		el.CreateText(prompt)
	}
	el := ib.NewXML("ExampleOutput")
	el.CreateText(`{"table_name":"user_info","table_description":"A table of basic user information","columns":[{"name":"name","description":"The full name of the user","type":"string"},{"name":"age","description":"The age of the user in years","type":"integer"},{"name":"job","description":"The occupation of the user","type":"string"}],"rows":[["Jack",24,"Doctor"],["Amy",36,"Teacher"]]}
`)
	ib.AddText("- table_name **must** start with a letter and contain only letters, numbers, or underscores.")
	return ib
}

func (ib *ImageToTableBuilder) AddExistingTableNames(names []string) {
	if len(names) > 0 {
		ib.AddText(fmt.Sprintf("- Avoid using table names that already exist: <tables>%s</tables>", strings.Join(names, ",")))
	}
}
