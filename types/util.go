package types

/*var builtins = map[]*/

func IsBuiltin(t Type) bool {
	if t.TypeOf() != T_Name {
		return false
	}
	return false
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
