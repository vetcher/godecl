package types

type Interface struct {
	Base
	Methods []*Function `json:"methods,omitempty"`
}
