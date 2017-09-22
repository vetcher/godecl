package types

import (
	"fmt"
	"strings"
)

type Interface struct {
	Base
	Methods []*Function `json:"methods,omitempty"`
}

func (i Interface) String() string {
	var methods []string
	for _, m := range i.Methods {
		methods = append(methods, m.funcStr())
	}
	return fmt.Sprintf("type %s interface {\n\t%s\n}", i.Name, strings.Join(methods, "\n\t"))
}

func (i Interface) GoString() string {
	return i.String()
}
