package types

type Function struct {
	Base
	Args    []Variable
	Results []Variable
}

type Method struct {
	Function
	Receiver Variable
}
