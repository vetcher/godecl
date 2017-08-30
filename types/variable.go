package types

type Variable struct {
	Base
	Type Type `json:"type,omitempty"`
}
