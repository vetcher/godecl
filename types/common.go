package types

type Documented interface {
	Comments() []string
}

type Base struct {
	Name string
	Docs []string
}
