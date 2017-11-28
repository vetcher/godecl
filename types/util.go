package types

var builtinTypes = map[string]bool{
	"bool":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
	"string":     true,
	"int":        true,
	"uint":       true,
	"uintptr":    true,
	"byte":       true,
	"rune":       true,
}

var builtinFunctions = map[string]bool{
	"append":  true,
	"copy":    true,
	"delete":  true,
	"len":     true,
	"cap":     true,
	"make":    true,
	"new":     true,
	"complex": true,
	"real":    true,
	"imag":    true,
	"close":   true,
	"panic":   true,
	"recover": true,
	"print":   true,
	"println": true,
}

func IsBuiltin(t Type) bool {
	if t.TypeOf() != T_Name {
		return false
	}
	for {
		switch tt := t.(type) {
		case TName:
			return IsBuiltinTypeString(tt.TypeName)
		default:
			next, ok := tt.(LinearType)
			if !ok {
				return false
			}
			t = next.NextType()
		}
	}
}

func IsBuiltinTypeString(t string) bool {
	return builtinTypes[t]
}

func IsBuiltinFuncString(t string) bool {
	return builtinFunctions[t]
}

func IsBuiltinString(t string) bool {
	return IsBuiltinTypeString(t) || IsBuiltinFuncString(t)
}

func TypeName(t Type) (string, bool) {
	for {
		switch tt := t.(type) {
		case TName:
			return tt.TypeName, true
		case TInterface:
			return "", false
		case TMap:
			return "", false
		default:
			next, ok := tt.(LinearType)
			if !ok {
				return "", false
			}
			t = next.NextType()
		}
	}
}

func TypeImport(t Type) *Import {
	for {
		switch tt := t.(type) {
		case TImport:
			return tt.Import
		default:
			next, ok := tt.(LinearType)
			if !ok {
				return nil
			}
			t = next.NextType()
		}
	}
}
