package types

import (
	"strconv"
	"strings"
)

type TypesOfTypes int32

const (
	T_Name TypesOfTypes = iota
	T_Pointer
	T_Array
	T_Map
	T_Interface
	T_Import
	T_Ellipsis
)

type Type interface {
	TypeOf() TypesOfTypes
	String() string
}

type LinearType interface {
	NextType() Type
}

type TInterface struct {
	Interface *Interface `json:"interface,omitempty"`
}

func (TInterface) TypeOf() TypesOfTypes {
	return T_Interface
}

func (i TInterface) String() string {
	if i.Interface != nil {
		return i.Interface.String()
	}
	return ""
}

type TMap struct {
	Key   Type `json:"key,omitempty"`
	Value Type `json:"value,omitempty"`
}

func (TMap) TypeOf() TypesOfTypes {
	return T_Map
}

func (m TMap) String() string {
	return "map[" + m.Key.String() + "]" + m.Value.String()
}

type TName struct {
	TypeName string `json:"type_name,omitempty"`
}

func (TName) TypeOf() TypesOfTypes {
	return T_Name
}

func (i TName) String() string {
	return i.TypeName
}

func (i TName) NextType() Type {
	return nil
}

type TPointer struct {
	NumberOfPointers int  `json:"number_of_pointers,omitempty"`
	Next             Type `json:"next,omitempty"`
}

func (TPointer) TypeOf() TypesOfTypes {
	return T_Pointer
}

func (i TPointer) String() string {
	str := strings.Repeat("*", i.NumberOfPointers)
	if i.Next != nil {
		str += i.Next.String()
	}
	return str
}

func (i TPointer) NextType() Type {
	return i.Next
}

type TArray struct {
	ArrayLen   int  `json:"array_len,omitempty"`
	IsSlice    bool `json:"is_slice,omitempty"` // [] declaration
	IsEllipsis bool `json:"is_ellipsis,omitempty"`
	Next       Type `json:"next,omitempty"`
}

func (TArray) TypeOf() TypesOfTypes {
	return T_Array
}

func (i TArray) String() string {
	str := ""
	if i.IsEllipsis {
		str += "..."
	} else if i.IsSlice {
		str += "[]"
	} else {
		str += "[" + strconv.Itoa(i.ArrayLen) + "]"
	}
	if i.Next != nil {
		str += i.Next.String()
	}
	return str
}

func (i TArray) NextType() Type {
	return i.Next
}

type TImport struct {
	Import *Import `json:"import,omitempty"`
	Next   Type    `json:"next,omitempty"`
}

func (TImport) TypeOf() TypesOfTypes {
	return T_Import
}

func (i TImport) String() string {
	str := ""
	if i.Import != nil {
		str += i.Import.Name
	}
	if i.Next != nil {
		str += i.Next.String()
	}
	return str
}

func (i TImport) NextType() Type {
	return i.Next
}

// TEllipsis used only for function params in declarations like `strs ...string`
type TEllipsis struct {
	Next Type `json:"next,omitempty"`
}

func (TEllipsis) TypeOf() TypesOfTypes {
	return T_Ellipsis
}

func (i TEllipsis) String() string {
	str := "..."
	if i.Next != nil {
		str += i.Next.String()
	}
	return str
}

func (i TEllipsis) NextType() Type {
	return i.Next
}

/*
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
*/
