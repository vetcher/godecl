package types

type StructField struct {
	Variable
	Tags    map[string][]string
	RawTags string
}

type Struct struct {
	Base
	Fields []StructField
}
