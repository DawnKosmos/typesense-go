package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// unwrapClientContainers removes the outer pointer from response maps and slices,
// including named container types. Element pointers and scalar/object pointers stay intact.
func unwrapClientContainers(clientPath, modelsPath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, clientPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	models, err := parser.ParseFile(fset, modelsPath, nil, 0)
	if err != nil {
		return err
	}
	types := map[string]ast.Expr{}
	for _, source := range []*ast.File{models, file} {
		for _, decl := range source.Decls {
			if declaration, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range declaration.Specs {
					if typ, ok := spec.(*ast.TypeSpec); ok {
						types[typ.Name.Name] = typ.Type
					}
				}
			}
		}
	}
	var container func(ast.Expr, map[string]bool) bool
	container = func(expr ast.Expr, seen map[string]bool) bool {
		switch typ := expr.(type) {
		case *ast.MapType, *ast.ArrayType:
			return true
		case *ast.Ident:
			if seen[typ.Name] {
				return false
			}
			seen[typ.Name] = true
			return container(types[typ.Name], seen)
		}
		return false
	}
	changed := map[string]map[string]ast.Expr{}
	for name, expr := range types {
		structure, ok := expr.(*ast.StructType)
		if !ok || !strings.HasSuffix(name, "Response") {
			continue
		}
		for _, field := range structure.Fields.List {
			pointer, ok := field.Type.(*ast.StarExpr)
			if !ok || !container(pointer.X, map[string]bool{}) {
				continue
			}
			for _, fieldName := range field.Names {
				if !strings.HasPrefix(fieldName.Name, "JSON") && !strings.HasPrefix(fieldName.Name, "XML") && !strings.HasPrefix(fieldName.Name, "YAML") {
					continue
				}
				if changed[name] == nil {
					changed[name] = map[string]ast.Expr{}
				}
				changed[name][fieldName.Name] = pointer.X
				field.Type = pointer.X
			}
		}
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Recv != nil && len(fn.Recv.List) == 1 {
			receiver := fn.Recv.List[0].Type
			if ptr, ok := receiver.(*ast.StarExpr); ok {
				receiver = ptr.X
			}
			if typ, ok := receiver.(*ast.Ident); ok && strings.HasPrefix(fn.Name.Name, "Get") {
				if fieldType := changed[typ.Name][strings.TrimPrefix(fn.Name.Name, "Get")]; fieldType != nil {
					fn.Type.Results.List[0].Type = fieldType
				}
			}
		}
		if !strings.HasPrefix(fn.Name.Name, "Parse") || fn.Type.Results == nil {
			continue
		}
		pointer, ok := fn.Type.Results.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		response, ok := pointer.X.(*ast.Ident)
		if !ok || len(changed[response.Name]) == 0 {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, lhs := range assignment.Lhs {
				field, ok := lhs.(*ast.SelectorExpr)
				if !ok || changed[response.Name][field.Sel.Name] == nil || i >= len(assignment.Rhs) {
					continue
				}
				if address, ok := assignment.Rhs[i].(*ast.UnaryExpr); ok && address.Op == token.AND {
					assignment.Rhs[i] = address.X
				}
			}
			return true
		})
	}
	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return err
	}
	return os.WriteFile(clientPath, output.Bytes(), 0644)
}
