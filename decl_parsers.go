package go_file_struct

import (
	"fmt"
	"go-file-struct/types"
	"go/ast"
	"go/token"
)

func parseDeclaration(decl ast.Decl, file *types.File) error {
	var onceVars map[string]struct{}
	switch d := decl.(type) {
	case *ast.GenDecl:
		switch d.Tok {
		case token.VAR:
			vars, err := parseVariables(d, file)
			if err != nil {
				return fmt.Errorf("parse variables %d:%d error: %v", d.Lparen, d.Rparen, err)
			}
			for _, variable := range vars {
				if _, ok := onceVars[variable.Name]; ok {
					return fmt.Errorf("duplicating variable %s", variable.Name)
				}
				onceVars[variable.Name] = struct{}{}
			}
			file.Vars = append(file.Vars, vars...)
		case token.TYPE:

		}
	}
	return nil
}

func parseVariables(decl *ast.GenDecl, file *types.File) (vars []types.Variable, err error) {
	spec := decl.Specs[0].(*ast.ValueSpec)
	if len(spec.Values) > 0 && len(spec.Values) != len(spec.Names) {
		err = fmt.Errorf("amount of variables and their values not same %d:%d", spec.Pos(), spec.End())
	}
	for i, name := range spec.Names {
		variable := types.Variable{
			Base: types.Base{
				Name: name.Name,
				Docs: parseComments(decl.Doc),
			},
		}
		var valType types.Type
		if spec.Type != nil {
			err := parseByType(&valType, spec.Type, file)
			if err != nil {
				return nil, err
			}
		} else {
			err := parseByValue(&valType, spec.Values[i], file)
			if err != nil {
				return nil, err
			}
		}

		variable.Type = valType
		vars = append(vars, variable)
	}
	return
}

func parseByType(tt *types.Type, spec interface{}, file *types.File) error {
	switch t := spec.(type) {
	case *ast.Ident:
		tt.Name = t.Name
	case *ast.SelectorExpr:
		tt.Name = t.Sel.Name
		tt.Import = findImportByAlias(file, t.X.(*ast.Ident).Name)
		tt.IsCustom = true
		if tt.Import == nil {
			return fmt.Errorf("wrong import %d:%d", t.Pos(), t.End())
		}
	case *ast.StarExpr:
		tt.IsPointer = true
		err := parseByType(tt, t.X, file)
		if err != nil {
			return err
		}
	case *ast.ArrayType:
		tt.IsArray = true
		err := parseByType(tt, t.Elt, file)
		if err != nil {
			return err
		}
	case *ast.MapType:
		var key, value types.Type
		err := parseByType(&key, t.Key, file)
		if err != nil {
			return err
		}
		err = parseByType(&value, t.Value, file)
		if err != nil {
			return err
		}
		tt.SetMap(key, value)
		tt.IsCustom = true
	default:
		return ErrUnexpectedSpec
	}
	return nil
}

func parseByValue(tt *types.Type, spec interface{}, file *types.File) error {
	switch t := spec.(type) {
	case *ast.BasicLit:
		tt.Name = t.Kind.String()
	case *ast.CompositeLit:
		return parseByValue(tt, t.Type, file)
	case *ast.SelectorExpr:
		tt.Name = t.Sel.Name
		tt.Import = findImportByAlias(file, t.X.(*ast.Ident).Name)
		tt.IsCustom = true
		if tt.Import == nil {
			return fmt.Errorf("wrong import %d:%d", t.Pos(), t.End())
		}
	}
	return nil
}
