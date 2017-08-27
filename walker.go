package go_file_struct

import (
	"errors"
	"go-file-struct/types"
	"go/ast"
	"go/token"
	"path"
	"strings"
)

var (
	ErrCouldNotResolvePackage  = errors.New("could not resolve package")
	ErrUnexpectedFieldType     = errors.New("provided fields have unexpected type")
	ErrNotUniquePackageAliases = errors.New("not unique package aliases")
	ErrUnexpectedSpec          = errors.New("unexpected spec")
)

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
	parseTopLevelDeclarations(file.Decls, f)
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

func parseTopLevelDeclarations(decls []ast.Decl, file *types.File) {
	for i := range decls {
		parseDeclaration(decls[i], file)
	}
}

func findImportByAlias(file *types.File, alias string) *types.Import {
	for _, imp := range file.Imports {
		if imp.Alias == alias {
			return &imp
		}
	}
	return nil
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
				//return nil, ErrUnexpectedFieldType
			}
			alias := constructAliasName(spec)
			if _, ok := onceImport[alias]; ok {
				return nil, ErrNotUniquePackageAliases
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
