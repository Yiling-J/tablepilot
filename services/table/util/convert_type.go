package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/spf13/cast"
)

func ZeroValue(tp tablecolumn.Type) (any, error) {
	switch tp {
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
	case tablecolumn.TypeImage:
		return "", nil
	default:
		return "", errors.New("unsupported type")
	}
}

func ConvertStringToType(v string, to tablecolumn.Type) (any, error) {
	if v == "" {
		return ZeroValue(to)
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
	case tablecolumn.TypeImage:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type: %v", v)
	}
}

func ConvertAnyToType(v any, to tablecolumn.Type) any {
	switch to {
	case tablecolumn.TypeString, tablecolumn.TypeImage:
		return cast.ToString(v)
	case tablecolumn.TypeNumber:
		return cast.ToFloat64(v)
	case tablecolumn.TypeInteger:
		return cast.ToInt(v)
	case tablecolumn.TypeBoolean:
		return cast.ToBool(v)
	case tablecolumn.TypeArray:
		switch vt := v.(type) {
		case string:
			var arr []any
			if err := json.Unmarshal([]byte(vt), &arr); err == nil {
				return arr
			}
		case []any:
			return v
		}
		return []any{}
	default:
		return cast.ToString(v)
	}
}
