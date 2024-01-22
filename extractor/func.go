package extractor

import (
	"go/ast"
	"strings"
)

type NameChangerWalker struct {
	alias        string
	pkgName      string
	receiverName string
	receiverType string
}

func NewNameChangerWalker(alias, pkgName, receiverName, receiverType string) *NameChangerWalker {
	return &NameChangerWalker{
		alias:        alias,
		pkgName:      pkgName,
		receiverName: receiverName,
		receiverType: receiverType,
	}
}

func (n *NameChangerWalker) Visit(i ast.Node) ast.Visitor {
	switch node := i.(type) {
	case *ast.Ident:
		if node.Name == n.receiverName {
			node.Name = n.alias
		}
		if node.Name == n.receiverType {
			node.Name = n.pkgName
		}
	}
	return n
}

type DestInterfaceScanner struct {
	Functions     []string
	InterfaceName string
}

func NewDestInterfaceScanner() *DestInterfaceScanner {
	return &DestInterfaceScanner{
		Functions: make([]string, 0),
	}
}

func (d *DestInterfaceScanner) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.InterfaceType:
		if m := node.Methods; m != nil {
			if len(m.List) != 0 {
				for _, v := range m.List {
					d.Functions = append(d.Functions, v.Names[0].Name)
				}
			}
		}
	case *ast.TypeSpec:
		d.InterfaceName = node.Name.Name
	}

	return d
}

type StructScanner struct {
	// Struct name to field name to field list
	Structs map[string]string
}

// Import Mapper
type ImportScanner struct {
	// Maps the Imports into path - name
	ByPath map[string]string

	// Maps the imports into name - path
	ByName map[string]string
}

func NewImportScanner() *ImportScanner {
	return &ImportScanner{
		ByPath: make(map[string]string),
		ByName: make(map[string]string),
	}
}

func (is *ImportScanner) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.ImportSpec:
		// Incase it doesnt have specific name get from last path
		fi := strings.Split(node.Path.Value, "/")
		if !(len(fi) > 0) {
			return is
		}

		name := ""
		if node.Name != nil {
			name = node.Name.Name
		}
		is.ByPath[cleanUpString(node.Path.Value)] = cleanUpString(name)

		if name == "" {
			name = fi[len(fi)-1]
		}
		is.ByName[cleanUpString(name)] = cleanUpString(node.Path.Value)
	}
	return is
}

func cleanUpString(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}

// FuncVisitor -- lists all function
type FuncVisitor struct {
	Functions map[string]ast.Node
}

func NewFunctionVisitor() *FuncVisitor {
	return &FuncVisitor{
		Functions: make(map[string]ast.Node),
	}
}

func (fv *FuncVisitor) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.FuncDecl:
		if !strings.Contains(node.Name.Name, "Test_") {
			fv.Functions[node.Name.Name] = node
		}
	}
	return fv
}
