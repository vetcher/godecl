package types

import (
	"fmt"
	"strconv"
)

// Type contains type information. It can not carry too complex types, such as `struct{func(map[interface{}]string)}`.
type Type struct {
	Name      string  `json:"name,omitempty"`
	Import    *Import `json:"import,omitempty"`     // Not nil if type imported.
	IsPointer bool    `json:"is_pointer,omitempty"` // True if type is pointer.
	IsArray   bool    `json:"is_array,omitempty"`   // True if type is array/slice.
	// Capacity of array.
	// 0 - array is slice or not a array at all.
	// -1 - ... founded in declaration.
	Len         int        `json:"len"`
	IsCustom    bool       `json:"is_custom,omitempty"`    // True if Import != nil or IsInterface == true.
	IsMap       bool       `json:"is_map,omitempty"`       // True if type is map.
	IsInterface bool       `json:"is_interface,omitempty"` // True if type is interface.
	Map         *MapType   // Field for carry map type.
	Interface   *Interface // Field for carry interface type.
}

func (t Type) String() string {
	str := ""
	if t.IsArray {
		str += "["
		switch t.Len {
		case 0:
			break
		case -1:
			str += "..."
		default:
			str += strconv.Itoa(t.Len)
		}
		str += "]"
	}
	if t.IsPointer {
		str += "*"
	}
	if t.IsMap {
		return str + fmt.Sprintf("map[%s]%s", t.Map.Key.String(), t.Map.Value.String())
	}
	if t.Import != nil {
		str += t.Import.Name + "."
	}
	return str + t.Name
}

func (t Type) GoString() string {
	return t.String()
}

type MapType struct {
	Key   Type
	Value Type
}

func NewMapType(key, value Type) *MapType {
	return &MapType{
		Key:   key,
		Value: value,
	}
}
