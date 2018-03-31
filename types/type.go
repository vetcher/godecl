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
	T_Chan
	T_Func
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
		str += i.Import.Name + "."
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
		return str + i.Next.String()
	}
	return str
}

func (i TEllipsis) NextType() Type {
	return i.Next
}

const (
	ChanDirSend = 1
	ChanDirRecv = 2
	ChanDirAny  = ChanDirSend | ChanDirRecv
)

type TChan struct {
	Direction int
	Next      Type
}

func (TChan) TypeOf() TypesOfTypes {
	return T_Chan
}

func (c TChan) NextType() Type {
	return c.Next
}

var strForChan = map[int]string{
	ChanDirSend: "chan<-",
	ChanDirRecv: "<-chan",
	ChanDirAny:  "chan",
}

func (c TChan) String() string {
	str := strForChan[c.Direction]
	if c.Next != nil {
		return str + " " + c.Next.String()
	}
	return str
}
