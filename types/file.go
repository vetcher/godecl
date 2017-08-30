package types

// File is a top-level entity, that contains all top-level declarations of the file.
type File struct {
	Base                   // `File.Name` is package name, `File.Docs` is a comments above `package ...`
	Imports    []Import    `json:"imports,omitempty"`    // Contains imports and their aliases from `import` blocks.
	Constants  []Variable  `json:"constants,omitempty"`  // Contains constant variables from `const` blocks.
	Vars       []Variable  `json:"vars,omitempty"`       // Contains variables from `var` blocks.
	Interfaces []Interface `json:"interfaces,omitempty"` // Contains `type Foo interface` declarations.
	Structures []Struct    `json:"structures,omitempty"` // Contains `type Foo struct` declarations.
	Functions  []Function  `json:"functions,omitempty"`  // Contains `func Foo() {}` declarations.
	Methods    []Method    `json:"methods,omitempty"`    // Contains `func (a A) Foo(b B) (c C) {}` declarations.
}
