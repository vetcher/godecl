package types

type Type struct {
	Name      string
	Import    *Import
	IsPointer bool
	IsArray   bool
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
