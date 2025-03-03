package util

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
)

// Item represents a single map entry with custom XML marshaling.
type Item struct {
	XMLName xml.Name
	Value   map[string]interface{}
}

// MarshalXML customizes the XML marshaling for the Item type.
func (i Item) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = i.XMLName
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for k, v := range i.Value {
		elem := xml.StartElement{Name: xml.Name{Local: k}}
		if err := e.EncodeElement(v, elem); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

type ColumnIndexer struct {
	columns  []*ent.TableColumn
	idMap    map[string]*ent.TableColumn
	orderMap map[string]int
}

func NewColumnIndexer(columns []*ent.TableColumn) *ColumnIndexer {
	ci := &ColumnIndexer{
		columns:  columns,
		idMap:    map[string]*ent.TableColumn{},
		orderMap: map[string]int{},
	}
	for i, col := range columns {
		ci.idMap[col.Nanoid] = col
		ci.orderMap[col.Nanoid] = i
	}
	return ci
}

func (ci *ColumnIndexer) GetColumnByIndex(index int) (*ent.TableColumn, error) {
	if index >= len(ci.columns) {
		return nil, errors.New("invalid index")
	}
	return ci.columns[index], nil
}

func (ci *ColumnIndexer) GetColumnByNanoid(id string) (*ent.TableColumn, error) {
	col, ok := ci.idMap[id]
	if !ok {
		return nil, errors.New("invalid id")
	}
	return col, nil
}

func (ci *ColumnIndexer) GetColumnIndexByNanoid(id string) (int, error) {
	index, ok := ci.orderMap[id]
	if !ok {
		return 0, errors.New("invalid id")
	}
	return index, nil
}

func (ci *ColumnIndexer) RowMapToSlice(row map[string]*schema.CellValue) ([]*schema.CellValue, error) {
	data := []*schema.CellValue{}
	for i := 0; i < len(ci.columns); i++ {
		col, err := ci.GetColumnByIndex(i)
		if err != nil {
			return nil, err
		}
		v, ok := row[col.Nanoid]
		if ok {
			data = append(data, v)
		} else {
			data = append(data, &schema.CellValue{})
		}
	}
	return data, nil
}

func (ci *ColumnIndexer) ColumnNames() []string {
	names := []string{}
	for _, col := range ci.columns {
		names = append(names, col.Name)
	}
	return names
}

// TryDecodeJsonArray attempts to decode a JSON array from a raw string into a slice of type T.
// If any elements fail to decode, it continues processing the remaining elements and returns
// the first encountered error (or joins multiple errors). The successfully decoded elements
// are still included in the result.
func TryDecodeJsonArray[T any](raw string) ([]T, error) {
	dec := json.NewDecoder(strings.NewReader(raw))
	result := []T{}

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, err
	}

	var decErr error
	for dec.More() {
		var m T
		err := dec.Decode(&m)
		if err != nil {
			if decErr == nil {
				decErr = err
			} else {
				decErr = errors.Join(decErr, err)
			}
		} else {
			result = append(result, m)
		}
	}

	return result, decErr
}
