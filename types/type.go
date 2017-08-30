package types

import (
	"fmt"
	"strconv"
)

type Type struct {
	Name      string
	Import    *Import
	IsPointer bool
	IsArray   bool
	Len       int
	IsCustom  bool
	IsMap     bool
	m         *mapType
}

func (t *Type) Map() *mapType {
	if t.IsMap {
		return t.m
	}
	panic("not a map type")
}

func (t *Type) String() string {
	if t.IsMap {
		return fmt.Sprintf("map[%s]%s", t.m.Key.String(), t.m.Value.String())
	}
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
	if t.Import != nil {
		str += t.Import.Alias + "."
	}
	return str + t.Name
}

func (t *Type) SetMap(key, value Type) {
	t.m = NewMapType(key, value)
	t.IsMap = true
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
