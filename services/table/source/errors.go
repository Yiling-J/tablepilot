package source

import "fmt"

func ErrTableNameOrIdEmpty() error {
	return &InvalidSource{
		s: "Table name or ID is empty",
	}
}

func ErrTableNotFound(input string) error {
	return &InvalidSource{
		s: fmt.Sprintf("Table not found: %s", input),
	}
}

func ErrColumnNameOrIdEmpty() error {
	return &InvalidSource{
		s: "Column name or ID is empty",
	}
}

func ErrColumnNotFound(input string) error {
	return &InvalidSource{
		s: fmt.Sprintf("Column not found: %s", input),
	}
}

type InvalidSource struct {
	s string
}

func (e *InvalidSource) Error() string {
	return e.s
}
