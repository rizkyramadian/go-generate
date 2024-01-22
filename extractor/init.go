package extractor

import "go/ast"

// Unused for now might user later to improve code
type extractor struct {
	// Destination
	PackageName string
	FuncFile    *ast.Node
	InitFile    *ast.Node
	Imports     []*ast.Node
	MainStruct  *ast.Node
}
