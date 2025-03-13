package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
)

func ConvertToType(v string, to tablecolumn.Type) (any, error) {
	if v == "" {
		switch to {
		case tablecolumn.TypeString:
			return "", nil
		case tablecolumn.TypeNumber:
			return 0.0, nil
		case tablecolumn.TypeInteger:
			return 0, nil
		case tablecolumn.TypeBoolean:
			return false, nil
		case tablecolumn.TypeArray:
			return []any{}, nil
		default:
			return nil, errors.New("unsupported type")
		}
	}

	switch to {
	case tablecolumn.TypeString:
		return v, nil
	case tablecolumn.TypeNumber:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, nil
		}
		return nil, fmt.Errorf("invalid number format: %v", v)
	case tablecolumn.TypeInteger:
		if num, err := strconv.Atoi(v); err == nil {
			return num, nil
		}
		return nil, fmt.Errorf("invalid integer format: %v", v)
	case tablecolumn.TypeBoolean:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, nil
		}
		return nil, fmt.Errorf("invalid boolean format: %v", v)
	case tablecolumn.TypeArray:
		var arr []any
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			return arr, nil
		}
		return nil, fmt.Errorf("invalid JSON array format: %v", v)
	default:
		return nil, fmt.Errorf("unsupported type: %v", v)
	}
}
