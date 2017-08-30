package types

type Function struct {
	Base
	Args    []Variable `json:"args,omitempty"`
	Results []Variable `json:"results,omitempty"`
}

type Method struct {
	Function
	Receiver Variable `json:"receiver,omitempty"`
}
