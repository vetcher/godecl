package types

// File is a top-level entity, that contains all top-level declarations of the file.
type File struct {
	Base                   // `File.Name` is package name, `File.Docs` is a comments above `package ...`
	Imports    []Import    // Contains imports and their aliases from `import` blocks.
	Constants  []Variable  // Contains constant variables from `const` blocks.
	Vars       []Variable  // Contains variables from `var` blocks.
	Interfaces []Interface // Contains `type Foo interface` declarations.
	Structures []Struct    // Contains `type Foo struct` declarations.
	Functions  []Function  // Contains `func Foo() {}` declarations.
	Methods    []Method    // Contains `func (a A) Foo(b B) (c C) {}` declarations.
}
