package godecl

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/vetcher/godecl/types"
)

var (
	ErrCouldNotResolvePackage = errors.New("could not resolve package")
	ErrNotUniquePackageAlias  = errors.New("not unique package alias")
	ErrUnexpectedSpec         = errors.New("unexpected spec")
)

// Parses ast.File and return all top-level declarations.
func ParseFile(file *ast.File) (*types.File, error) {
	f := &types.File{
		Base: types.Base{
			Name: file.Name.Name,
			Docs: parseComments(file.Doc),
		},
	}
	imports, err := parseImports(file)
	if err != nil {
		return nil, err
	}
	f.Imports = imports
	err = parseTopLevelDeclarations(file.Decls, f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func parseComments(group *ast.CommentGroup) (comments []string) {
	if group == nil {
		return
	}
	for _, comment := range group.List {
		comments = append(comments, comment.Text)
	}
	return
}

func parseTopLevelDeclarations(decls []ast.Decl, file *types.File) error {
	for i := range decls {
		err := parseDeclaration(decls[i], file)
		if err != nil {
			return err
		}
	}
	return nil
}

func findImportByAlias(file *types.File, alias string) (*types.Import, error) {
	for _, imp := range file.Imports {
		if imp.Alias == alias {
			return &imp, nil
		}
	}
	// try to find by last package path
	for _, imp := range file.Imports {
		if alias == path.Base(imp.Package) {
			return &imp, nil
		}
	}

	return nil, fmt.Errorf("%v: %s", ErrCouldNotResolvePackage, alias)
}

func parseImports(f *ast.File) ([]types.Import, error) {
	for _, decl := range f.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.IMPORT {
			continue
		}

		var imports []types.Import
		var onceImport map[string]struct{}

		for _, spec := range decl.Specs {
			spec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue // if !ok then comment
			}
			alias := constructAliasName(spec)
			if _, ok := onceImport[alias]; ok {
				return nil, fmt.Errorf("%v: %s", ErrNotUniquePackageAlias, alias)
			}

			imp := types.Import{
				Alias:   alias,
				Package: strings.Trim(spec.Path.Value, `"`),
				Docs:    parseComments(spec.Doc),
			}

			imports = append(imports, imp)
		}

		return imports, nil
	}

	return nil, nil
}

func constructAliasName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	return path.Base(strings.Trim(spec.Path.Value, `"`))
}

func parseDeclaration(decl ast.Decl, file *types.File) error {
	onceVars := make(map[string]struct{})
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
		case token.CONST:
			consts, err := parseVariables(d, file)
			if err != nil {
				return fmt.Errorf("parse variables %d:%d error: %v", d.Lparen, d.Rparen, err)
			}
			for _, variable := range consts {
				if _, ok := onceVars[variable.Name]; ok {
					return fmt.Errorf("duplicating variable %s", variable.Name)
				}
				onceVars[variable.Name] = struct{}{}
			}
			file.Constants = append(file.Constants, consts...)
		case token.TYPE:
			typeSpec := d.Specs[0].(*ast.TypeSpec)
			switch t := typeSpec.Type.(type) {
			case *ast.InterfaceType:
				methods, err := parseInterfaceMethods(t, file)
				if err != nil {
					return err
				}
				file.Interfaces = append(file.Interfaces, types.Interface{
					Base: types.Base{
						Name: typeSpec.Name.Name,
						Docs: parseComments(d.Doc),
					},
					Methods: methods,
				})
			case *ast.StructType:
				strFields, err := parseStructFields(t, file)
				if err != nil {
					return fmt.Errorf("%s: can't parse struct fields: %v", typeSpec.Name.Name, err)
				}
				file.Structures = append(file.Structures, types.Struct{
					Base: types.Base{
						Name: typeSpec.Name.Name,
						Docs: parseComments(d.Doc),
					},
					Fields: strFields,
				})
			}
		}
	case *ast.FuncDecl:
		fn := types.Function{
			Base: types.Base{
				Name: d.Name.Name,
				Docs: parseComments(d.Doc),
			},
		}
		fmt.Println("YEY")
		err := parseFuncParamsAndResults(d.Type, &fn, file)
		if err != nil {
			return fmt.Errorf("parse func %s error: %v", fn.Name, err)
		}
		if d.Recv != nil {
			rec, err := parseReceiver(d.Recv, file)
			if err != nil {
				return err
			}
			file.Methods = append(file.Methods, types.Method{
				Function: fn,
				Receiver: rec,
			})
		} else {
			file.Functions = append(file.Functions, fn)
		}
		fmt.Println(len(file.Functions))
	}
	return nil
}

func parseReceiver(list *ast.FieldList, file *types.File) (t types.Variable, err error) {
	recv, err := parseParams(list, file)
	if err != nil {
		return
	}
	if len(recv) != 0 {
		return recv[0], nil
	}
	err = fmt.Errorf("reciever not found for %d:%d", list.Pos(), list.End())
	return
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
				return nil, fmt.Errorf("can't parse type: %v", err)
			}
		} else {
			err := parseByValue(&valType, spec.Values[i], file)
			if err != nil {
				return nil, fmt.Errorf("can't parse type: %v", err)
			}
		}

		variable.Type = valType
		vars = append(vars, variable)
	}
	return
}

// Fill provided types.Type for cases, when variable's type is provided.
func parseByType(tt *types.Type, spec interface{}, file *types.File) (err error) {
	switch t := spec.(type) {
	case *ast.Ident:
		tt.Name = t.Name
	case *ast.SelectorExpr:
		tt.Name = t.Sel.Name
		tt.Import, err = findImportByAlias(file, t.X.(*ast.Ident).Name)
		if err != nil {
			return fmt.Errorf("%s: %v", tt.Name, err)
		}
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
		tt.Len = parseArrayLen(t)
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
	case *ast.InterfaceType:
		methods, err := parseInterfaceMethods(t, file)
		if err != nil {
			return err
		}
		tt.IsCustom = true
		tt.SetInterface(types.Interface{
			Base: types.Base{
				Name: tt.Name,
			},
			Methods: methods,
		})
	default:
		return ErrUnexpectedSpec
	}
	return nil
}

func parseArrayLen(t *ast.ArrayType) int {
	if t == nil {
		return 0
	}
	switch l := t.Len.(type) {
	case *ast.Ellipsis:
		return -1
	case *ast.BasicLit:
		if l.Kind == token.INT {
			x, _ := strconv.Atoi(l.Value)
			return x
		}
		return 0
	}
	return 0
}

// Fill provided types.Type for cases, when variable's value is provided.
func parseByValue(tt *types.Type, spec interface{}, file *types.File) (err error) {
	switch t := spec.(type) {
	case *ast.BasicLit:
		tt.Name = t.Kind.String()
	case *ast.CompositeLit:
		return parseByValue(tt, t.Type, file)
	case *ast.SelectorExpr:
		tt.Name = t.Sel.Name
		tt.Import, err = findImportByAlias(file, t.X.(*ast.Ident).Name)
		if err != nil {
			return fmt.Errorf("%s: %v", tt.Name, err)
		}
		tt.IsCustom = true
		if tt.Import == nil {
			return fmt.Errorf("wrong import %d:%d", t.Pos(), t.End())
		}
	}
	return nil
}

// Collects and returns all interface methods.
func parseInterfaceMethods(ifaceType *ast.InterfaceType, file *types.File) ([]*types.Function, error) {
	var fns []*types.Function
	if ifaceType.Methods != nil {
		for _, method := range ifaceType.Methods.List {
			fn, err := parseFunction(method, file)
			if err != nil {
				return nil, err
			}
			fns = append(fns, fn)
		}
	}
	return fns, nil
}

func parseFunction(funcField *ast.Field, file *types.File) (*types.Function, error) {
	fn := &types.Function{
		Base: types.Base{
			Name: funcField.Names[0].Name,
			Docs: parseComments(funcField.Doc),
		},
	}
	funcType := funcField.Type.(*ast.FuncType)
	err := parseFuncParamsAndResults(funcType, fn, file)
	if err != nil {
		return nil, err
	}
	return fn, nil
}

func parseFuncParamsAndResults(funcType *ast.FuncType, fn *types.Function, file *types.File) error {
	args, err := parseParams(funcType.Params, file)
	if err != nil {
		return fmt.Errorf("can't parse args: %v", err)
	}
	fn.Args = args
	results, err := parseParams(funcType.Results, file)
	if err != nil {
		return fmt.Errorf("can't parse results: %v", err)
	}
	fn.Results = results
	return nil
}

// Collects and returns all args/results from function or fields from structure.
func parseParams(fields *ast.FieldList, file *types.File) ([]types.Variable, error) {
	var vars []types.Variable
	if fields != nil {
		for _, field := range fields.List {
			if field.Type == nil {
				return nil, fmt.Errorf("param's type is nil %d:%d", field.Pos(), field.End())
			}
			t := types.Type{}
			err := parseByType(&t, field.Type, file)
			if err != nil {
				return nil, fmt.Errorf("wrong type: %v", err)
			}
			docs := parseComments(field.Doc)
			if len(field.Names) == 0 {
				vars = append(vars, types.Variable{
					Base: types.Base{
						Docs: docs,
					},
					Type: t,
				})
			} else {
				for _, name := range field.Names {
					vars = append(vars, types.Variable{
						Base: types.Base{
							Name: name.Name,
							Docs: docs,
						},
						Type: t,
					})
				}
			}
		}
	}
	return vars, nil
}

func parseTags(lit *ast.BasicLit) (tags map[string][]string, raw string) {
	if lit == nil {
		return
	}
	raw = lit.Value
	str := strings.Trim(lit.Value, "`")
	t, err := structtag.Parse(str)
	if err != nil {
		return
	}
	tags = make(map[string][]string)
	for _, tag := range t.Tags() {
		tags[tag.Key] = append([]string{tag.Name}, tag.Options...)
	}
	return
}

func parseStructFields(s *ast.StructType, file *types.File) ([]types.StructField, error) {
	fields, err := parseParams(s.Fields, file)
	if err != nil {
		return nil, err
	}
	var strF []types.StructField
	for i, f := range fields {
		parsedTags, rawTags := parseTags(s.Fields.List[i].Tag)
		strF = append(strF, types.StructField{
			Variable: f,
			Tags:     parsedTags,
			RawTags:  rawTags,
		})
	}
	return strF, nil
}
