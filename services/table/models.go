package table

import (
	"encoding/json"

	"github.com/Yiling-J/tablepilot/ent"
)

type TableGenColumn struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	FillMode      string `json:"fill_mode"`
	Source        string `json:"source"`
	Repeat        int    `json:"repeat"`
	ContextLength int    `json:"context_length"`
}

type TableGenRequest struct {
	Name        string            `json:"name"`
	Model       string            `json:"model"`
	Description string            `json:"description"`
	Columns     []TableGenColumn  `json:"columns"`
	Sources     []json.RawMessage `json:"sources"`
}

type GenerateRowsRequest struct {
	Table       string
	SaveTo      string
	Count       int
	Batch       int
	Temperature float64
	Model       string
}

type ColumnSchema struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	LinkedTo string `json:"linked_to"`
	Mode     string `json:"mode"`
}

type TableInfoSimple struct {
	ID          string
	Name        string
	Description string
	Model       string
}

type ListTablesResponse struct {
	Total  int
	Tables []TableInfoSimple
}

type TableColumnInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	FillMode    string `json:"fill_mode"`
}

type Rows struct {
	Columns []*ent.TableColumn
	Rows    []*ent.TableRow
}
