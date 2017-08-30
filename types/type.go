package types

import (
	"fmt"
	"strconv"
)

// Type contains type information. It can not carry too complex types, such as `struct{func(map[interface{}]string)}`.
type Type struct {
	Name      string
	Import    *Import
	IsPointer bool // True if type is pointer.
	IsArray   bool // True if type is array/slice.
	// Capacity of array.
	// 0 - array is slice or not a array at all.
	// -1 - ... founded in declaration.
	Len         int
	IsCustom    bool       // True if Import != nil or type is interface.
	IsMap       bool       // True if type is map.
	IsInterface bool       // True if type is interface.
	m           *mapType   // Hided field for carry map type. Use Map() to access.
	i           *Interface // Hided field for carry interface type. Use Interface() to access.
}

func (t *Type) Map() *mapType {
	if t.IsMap {
		return t.m
	}
	panic("not a map type")
}

func (t *Type) String() string {
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
		return str + fmt.Sprintf("map[%s]%s", t.m.Key.String(), t.m.Value.String())
	}
	if t.Import != nil {
		str += t.Import.Alias + "."
	}
	return str + t.Name
}

func (t *Type) SetMap(key, value Type) {
	t.m = NewMapType(key, value)
	t.IsMap = true
}

func (t *Type) SetInterface(iface Interface) {
	i := iface
	t.i = &i
	t.IsInterface = true
}

func (t *Type) Interface() *Interface {
	if t.IsInterface {
		return t.i
	}
	panic("not a interface type")
}

type mapType struct {
	Key   Type
	Value Type
}

func NewMapType(key, value Type) *mapType {
	return &mapType{
		Key:   key,
		Value: value,
	}
}
