package types

type StructField struct {
	Variable
	Tags    map[string][]string
	RawTags string // Raw string from source.
}

type Struct struct {
	Base
	Fields []StructField
}
