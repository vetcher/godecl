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
	ErrUnexpectedSpec         = errors.New("unexpected spec")
	ErrNotInGoPath            = errors.New("not in GOPATH")
	ErrGoPathIsEmpty          = errors.New("GOPATH is empty")
)

type Option uint

const (
	IgnoreComments Option = 1
	IgnoreStructs  Option = iota * 2
	IgnoreInterfaces
)

func concatOptions(ops []Option) (o Option) {
	for i := range ops {
		o |= ops[i]
	}
	return
}

func (o Option) check(what Option) bool {
	return o&what == what
}

// Parses ast.File and return all top-level declarations.
func ParseAstFile(file *ast.File, packagePath string, options ...Option) (*types.File, error) {
	opt := concatOptions(options)
	f := &types.File{
		Base: types.Base{
			Name: file.Name.Name,
			Docs: parseComments(file.Doc, opt),
		},
	}
	var pp *types.Import
	if packagePath != "" {
		alias := constructAliasNameString(packagePath)
		imp := &types.Import{
			Base: types.Base{
				Name: alias,
			},
			Package: strings.Trim(packagePath, `"`),
		}
		f.Imports = append(f.Imports, imp)
		pp = imp
	}
	err := parseTopLevelDeclarations(file.Decls, f, pp)
	if err != nil {
		return nil, err
	}
	err = linkMethodsToStructs(f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func linkMethodsToStructs(f *types.File) error {
	for i := range f.Methods {
		structure, err := findStructByMethod(f, &f.Methods[i])
		if err != nil {
			return err
		}
		if structure != nil {
			structure.Methods = append(structure.Methods, &f.Methods[i])
			continue
		}
		typee, err := findTypeByMethod(f, &f.Methods[i])
		if err != nil {
			return err
		}
		if typee != nil {
			typee.Methods = append(typee.Methods, &f.Methods[i])
			continue
		}
	}
	return nil
}

func parseComments(group *ast.CommentGroup, o Option) (comments []string) {
	if o.check(IgnoreComments) {
		return
	}
	if group == nil {
		return
	}
	for _, comment := range group.List {
		comments = append(comments, comment.Text)
	}
	return
}

func parseTopLevelDeclarations(decls []ast.Decl, file *types.File, pp *types.Import) error {
	for i := range decls {
		err := parseDeclaration(decls[i], file, pp)
		if err != nil {
			return err
		}
	}
	return nil
}

func constructAliasName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	return constructAliasNameString(spec.Path.Value)
}

func constructAliasNameString(str string) string {
	return path.Base(strings.Trim(str, `"`))
}

func parseDeclaration(decl ast.Decl, file *types.File, pp *types.Import) error {
	switch d := decl.(type) {
	case *ast.GenDecl:
		switch d.Tok {
		case token.IMPORT:
			var imports []*types.Import

			for _, spec := range d.Specs {
				spec, ok := spec.(*ast.ImportSpec)
				if !ok {
					continue // if !ok then comment
				}
				alias := constructAliasName(spec)
				imp := &types.Import{
					Base: types.Base{
						Name: alias,
						Docs: parseComments(spec.Doc, 0),
					},
					Package: strings.Trim(spec.Path.Value, `"`),
				}

				imports = append(imports, imp)
			}
			file.Imports = append(file.Imports, imports...)
		case token.VAR:
			vars, err := parseVariables(d, file, pp)
			if err != nil {
				return fmt.Errorf("parse variables %d:%d error: %v", d.Lparen, d.Rparen, err)
			}
			file.Vars = append(file.Vars, vars...)
		case token.CONST:
			consts, err := parseVariables(d, file, pp)
			if err != nil {
				return fmt.Errorf("parse constants %d:%d error: %v", d.Lparen, d.Rparen, err)
			}
			file.Constants = append(file.Constants, consts...)
		case token.TYPE:
			typeSpec := d.Specs[0].(*ast.TypeSpec)
			switch t := typeSpec.Type.(type) {
			case *ast.InterfaceType:
				methods, err := parseInterfaceMethods(t, file, pp)
				if err != nil {
					return err
				}
				file.Interfaces = append(file.Interfaces, types.Interface{
					Base: types.Base{
						Name: typeSpec.Name.Name,
						Docs: parseComments(d.Doc, 0),
					},
					Methods: methods,
				})
			case *ast.StructType:
				strFields, err := parseStructFields(t, file, pp)
				if err != nil {
					return fmt.Errorf("%s: can't parse struct fields: %v", typeSpec.Name.Name, err)
				}
				file.Structures = append(file.Structures, types.Struct{
					Base: types.Base{
						Name: typeSpec.Name.Name,
						Docs: parseComments(d.Doc, 0),
					},
					Fields: strFields,
				})
			default:
				newType, err := parseByType(typeSpec.Type, file, pp)
				if err != nil {
					return fmt.Errorf("%s: can't parse type: %v", typeSpec.Name.Name, err)
				}
				file.Types = append(file.Types, types.FileType{Base: types.Base{
					Name: typeSpec.Name.Name,
					Docs: parseComments(d.Doc, 0),
				}, Type: newType})
			}
		}
	case *ast.FuncDecl:
		fn := types.Function{
			Base: types.Base{
				Name: d.Name.Name,
				Docs: parseComments(d.Doc, 0),
			},
		}
		err := parseFuncParamsAndResults(d.Type, &fn, file, pp)
		if err != nil {
			return fmt.Errorf("parse func %s error: %v", fn.Name, err)
		}
		if d.Recv != nil {
			rec, err := parseReceiver(d.Recv, file, pp)
			if err != nil {
				return err
			}
			file.Methods = append(file.Methods, types.Method{
				Function: fn,
				Receiver: *rec,
			})
		} else {
			file.Functions = append(file.Functions, fn)
		}
	}
	return nil
}

func parseReceiver(list *ast.FieldList, file *types.File, pp *types.Import) (*types.Variable, error) {
	recv, err := parseParams(list, file, pp)
	if err != nil {
		return nil, err
	}
	if len(recv) != 0 {
		return &recv[0], nil
	}
	return nil, fmt.Errorf("reciever not found for %d:%d", list.Pos(), list.End())
}

func parseVariables(decl *ast.GenDecl, file *types.File, pp *types.Import) (vars []types.Variable, err error) {
	spec := decl.Specs[0].(*ast.ValueSpec)
	if len(spec.Values) > 0 && len(spec.Values) != len(spec.Names) {
		return nil, fmt.Errorf("amount of variables and their values not same %d:%d", spec.Pos(), spec.End())
	}
	for i, name := range spec.Names {
		variable := types.Variable{
			Base: types.Base{
				Name: name.Name,
				Docs: parseComments(decl.Doc, 0),
			},
		}
		var (
			valType types.Type
			err     error
		)
		if spec.Type != nil {
			valType, err = parseByType(spec.Type, file, pp)
			if err != nil {
				return nil, fmt.Errorf("can't parse type: %v", err)
			}
		} else {
			valType, err = parseByValue(spec.Values[i], file)
			if err != nil {
				return nil, fmt.Errorf("can't parse type: %v", err)
			}
		}

		variable.Type = valType
		vars = append(vars, variable)
	}
	return
}

func setupImportIfNeed(tt types.Type, tImport *types.Import) types.Type {
	if tImport == nil {
		return tt
	}
	if types.IsBuiltinTypeString(tt.String()) {
		return tt
	}
	return types.TImport{Import: tImport, Next: tt}
}

// Fill provided types.Type for cases, when variable's type is provided.
func parseByType(spec interface{}, file *types.File, pp *types.Import) (tt types.Type, err error) {
	switch t := spec.(type) {
	case *ast.Ident:
		return types.TName{TypeName: t.Name}, nil
	case *ast.SelectorExpr:
		im, err := findImportByAlias(file, t.X.(*ast.Ident).Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", t.Sel.Name, err)
		}
		if im == nil {
			return nil, fmt.Errorf("wrong import %d:%d", t.Pos(), t.End())
		}
		return types.TImport{Import: im, Next: types.TName{TypeName: t.Sel.Name}}, nil
	case *ast.StarExpr:
		next, err := parseByType(t.X, file, pp)
		if err != nil {
			return nil, err
		}
		if next.TypeOf() == types.T_Pointer {
			return types.TPointer{Next: next.(types.TPointer).NextType(), NumberOfPointers: 1 + next.(types.TPointer).NumberOfPointers}, nil
		}
		return types.TPointer{Next: next, NumberOfPointers: 1}, nil
	case *ast.ArrayType:
		l := parseArrayLen(t)
		next, err := parseByType(t.Elt, file, pp)
		if err != nil {
			return nil, err
		}
		switch l {
		case -3, -2:
			return types.TArray{Next: next, IsSlice: true}, nil
		case -1:
			return types.TArray{Next: next, IsEllipsis: true}, nil
		default:
			return types.TArray{Next: next, ArrayLen: l}, nil
		}
	case *ast.MapType:
		key, err := parseByType(t.Key, file, pp)
		if err != nil {
			return nil, err
		}
		value, err := parseByType(t.Value, file, pp)
		if err != nil {
			return nil, err
		}
		return types.TMap{Key: key, Value: value}, nil
	case *ast.InterfaceType:
		methods, err := parseInterfaceMethods(t, file, pp)
		if err != nil {
			return nil, err
		}
		return types.TInterface{
			Interface: &types.Interface{
				Base:    types.Base{},
				Methods: methods,
			},
		}, nil
	case *ast.Ellipsis:
		next, err := parseByType(t.Elt, file, pp)
		if err != nil {
			return nil, err
		}
		return types.TEllipsis{Next: next}, nil
	case *ast.ChanType:
		next, err := parseByType(t.Value, file, pp)
		if err != nil {
			return nil, err
		}
		return types.TChan{Next: next, Direction: int(t.Dir)}, nil
	case *ast.ParenExpr:
		return parseByType(t.X, file, pp)
	case *ast.BadExpr:
		return nil, fmt.Errorf("bad expression")
	default:
		return nil, fmt.Errorf("%v: %T", ErrUnexpectedSpec, t)
	}
}

func parseArrayLen(t *ast.ArrayType) int {
	if t == nil {
		return -2
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
	return -3
}

// Fill provided types.Type for cases, when variable's value is provided.
func parseByValue(spec interface{}, file *types.File) (tt types.Type, err error) {
	switch t := spec.(type) {
	case *ast.BasicLit:
		return types.TName{TypeName: t.Kind.String()}, nil
	case *ast.CompositeLit:
		return parseByValue(t.Type, file)
	case *ast.SelectorExpr:
		im, err := findImportByAlias(file, t.X.(*ast.Ident).Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", t.Sel.Name, err)
		}
		if im == nil {
			return nil, fmt.Errorf("wrong import %d:%d", t.Pos(), t.End())
		}
		return types.TImport{Import: im}, nil
	default:
		return nil, nil
	}
}

// Collects and returns all interface methods.
func parseInterfaceMethods(ifaceType *ast.InterfaceType, file *types.File, pp *types.Import) ([]*types.Function, error) {
	var fns []*types.Function
	if ifaceType.Methods != nil {
		for _, method := range ifaceType.Methods.List {
			fn, err := parseFunction(method, file, pp)
			if err != nil {
				return nil, err
			}
			fns = append(fns, fn)
		}
	}
	return fns, nil
}

func parseFunction(funcField *ast.Field, file *types.File, pp *types.Import) (*types.Function, error) {
	fn := &types.Function{
		Base: types.Base{
			Name: funcField.Names[0].Name,
			Docs: parseComments(funcField.Doc, 0),
		},
	}
	funcType := funcField.Type.(*ast.FuncType)
	err := parseFuncParamsAndResults(funcType, fn, file, pp)
	if err != nil {
		return nil, err
	}
	return fn, nil
}

func parseFuncParamsAndResults(funcType *ast.FuncType, fn *types.Function, file *types.File, pp *types.Import) error {
	args, err := parseParams(funcType.Params, file, pp)
	if err != nil {
		return fmt.Errorf("can't parse args: %v", err)
	}
	fn.Args = args
	results, err := parseParams(funcType.Results, file, pp)
	if err != nil {
		return fmt.Errorf("can't parse results: %v", err)
	}
	fn.Results = results
	return nil
}

// Collects and returns all args/results from function or fields from structure.
func parseParams(fields *ast.FieldList, file *types.File, pp *types.Import) ([]types.Variable, error) {
	var vars []types.Variable
	if fields == nil {
		return vars, nil
	}
	for _, field := range fields.List {
		if field.Type == nil {
			return nil, fmt.Errorf("param's type is nil %d:%d", field.Pos(), field.End())
		}
		t, err := parseByType(field.Type, file, pp)
		if err != nil {
			return nil, fmt.Errorf("wrong type of %s: %v", strings.Join(namesOfIdents(field.Names), ","), err)
		}
		docs := parseComments(field.Doc, 0)
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

func parseStructFields(s *ast.StructType, file *types.File, pp *types.Import) ([]types.StructField, error) {
	fields, err := parseParams(s.Fields, file, pp)
	if err != nil {
		return nil, err
	}
	var strF []types.StructField
	for i, f := range fields {
		var tags *ast.BasicLit
		// Fill tags, if Tag field exist in ast
		if i < len(s.Fields.List) {
			tags = s.Fields.List[i].Tag
		}
		parsedTags, rawTags := parseTags(tags)
		strF = append(strF, types.StructField{
			Variable: f,
			Tags:     parsedTags,
			RawTags:  rawTags,
		})
	}
	return strF, nil
}

func findImportByAlias(file *types.File, alias string) (*types.Import, error) {
	for _, imp := range file.Imports {
		if imp.Name == alias {
			return imp, nil
		}
	}
	// try to find by last package path
	for _, imp := range file.Imports {
		if alias == path.Base(imp.Package) {
			return imp, nil
		}
	}

	return nil, fmt.Errorf("%v: %s", ErrCouldNotResolvePackage, alias)
}

func findStructByMethod(file *types.File, method *types.Method) (*types.Struct, error) {
	recType := method.Receiver.Type
	if !IsCommonReciever(recType) {
		return nil, fmt.Errorf("%s has not common reciever", method.String())
	}
	name := types.TypeName(recType)
	if name == nil {
		return nil, nil
	}
	for i := range file.Structures {
		if file.Structures[i].Name == *name {
			return &file.Structures[i], nil
		}
	}
	return nil, nil
}

func findTypeByMethod(file *types.File, method *types.Method) (*types.FileType, error) {
	recType := method.Receiver.Type
	if !IsCommonReciever(recType) {
		return nil, fmt.Errorf("%s has not common reciever", method.String())
	}
	name := types.TypeName(recType)
	if name == nil {
		return nil, nil
	}
	for i := range file.Types {
		if file.Types[i].Name == *name {
			return &file.Types[i], nil
		}
	}
	return nil, nil
}

func IsCommonReciever(t types.Type) bool {
	for tt := t; tt != nil; {
		switch tt.TypeOf() {
		case types.T_Array, types.T_Interface, types.T_Map, types.T_Import:
			return false
		default:
			x, ok := tt.(types.LinearType)
			if !ok {
				return false
			}
			tt = x.NextType()
			continue
		}
	}
	return true
}
