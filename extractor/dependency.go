package extractor

import (
	"go/ast"
	"strings"

	"github.com/rizkyramadian/go-generate/utility/hierarchy"
)

type selfDepCall struct {
	function  string
	selectors []string
}

// selfDepChecker runs the self dependency checks and creates array of string to be printed
func SelfDepChecker(node ast.Node) *hierarchy.StringNode {
	fnD, ok := node.(*ast.FuncDecl)
	if !ok {
		return nil
	}
	// res := make([]string, 0, 100)
	// Check For Receiver
	rn := receiverName(fnD)
	// rt := receiverType(fnD)

	// Process Function Declaration
	results := make([]*selfDepCall, 0)
	for _, v := range fnD.Body.List {
		results = selfDepStmtStrategy(v, results)
	}

	top := selfDepHierarchyGrouping(results, rn)
	return top
}

func selfDepTraverseAndFormat(node *hierarchy.StringNode, depth int, res []string, rn, rt string) []string {
	if node == nil {
		return []string{}
	}

	var str string
	str += strings.Repeat("   ", depth+1)
	str += "┗━ "
	str += node.Value
	if node.Value == rn {
		str += " (" + rt + ")"
	}
	res = append(res, str)
	depth++
	for _, v := range node.Children {
		res = selfDepTraverseAndFormat(v, depth, res, rn, rt)
	}
	return res
}

func selfDepHierarchyGrouping(list []*selfDepCall, rn string) *hierarchy.StringNode {
	// if Receiver name is "" that means the function does not have any self dependency
	if rn == "" {
		return nil
	}

	// Create new top level node
	topNode := hierarchy.NewStringHierarchy(rn)

	for _, v := range list {
		if len(v.selectors) < 1 {
			continue
		}
		// if the selector itself is different than receiver name can just skip it
		if v.selectors[0] != rn {
			continue
		}

		// Skip first selector because the first selector would be the top level
		// Generate Tree Nodes beforehand
		generateTreeNodes(topNode, v.selectors[1:])

		// Add Function to the furthest child
		node, _ := getFurthestChildByValue(topNode, v.selectors[1:], 0)
		// Prepare String
		var str string
		str += v.function + "()"
		node.AddChild(&hierarchy.StringNode{
			Value: str,
		})
	}

	return topNode
}

func generateTreeNodes(node *hierarchy.StringNode, branches []string) {
	// Get the furthest child based on its node
	child, idx := getFurthestChildByValue(node, branches, 0)
	// if child is nil means child hasnt been created
	if child == nil {
		// Safeguard check
		new := node.FindChild(branches[idx])
		if new == nil {
			new = node.AddChild(&hierarchy.StringNode{
				Value: branches[idx],
			})
		}
		generateTreeNodes(new, branches[1:])
	}
}

func getFurthestChildByValue(node *hierarchy.StringNode, values []string, idx int) (*hierarchy.StringNode, int) {
	// This should show furthest child
	if len(values) < 1 {
		return node, idx
	}

	current := node.FindChild(values[0])
	if current == nil {
		return nil, idx
	}
	return getFurthestChildByValue(node.FindChild(values[0]), values[1:], idx)
}

// Functions for Self Dependency Checkers
func selfDepStmtStrategy(stmt ast.Stmt, res []*selfDepCall) []*selfDepCall {
	switch stmt := stmt.(type) {
	case *ast.DeferStmt:
		res = selfDepExprStrategy(stmt.Call, res)
	case *ast.BadStmt:
		// Do Nothing
	case *ast.AssignStmt:
		for _, v := range stmt.Rhs {
			res = selfDepExprStrategy(v, res)
		}
	case *ast.RangeStmt:
		res = selfDepExprStrategy(stmt.Key, res)
		res = selfDepExprStrategy(stmt.Value, res)
		res = selfDepExprStrategy(stmt.X, res)
		res = selfDepStmtStrategy(stmt.Body, res)
	case *ast.ForStmt:
		res = selfDepStmtStrategy(stmt.Init, res)
		res = selfDepStmtStrategy(stmt.Post, res)
		res = selfDepExprStrategy(stmt.Cond, res)
		res = selfDepStmtStrategy(stmt.Body, res)
	case *ast.GoStmt:
		res = selfDepExprStrategy(stmt.Call, res)
	case *ast.IfStmt:
		res = selfDepStmtStrategy(stmt.Init, res)
		res = selfDepExprStrategy(stmt.Cond, res)
		res = selfDepStmtStrategy(stmt.Body, res)
		res = selfDepStmtStrategy(stmt.Else, res)
	case *ast.BlockStmt:
		for _, v := range stmt.List {
			res = selfDepStmtStrategy(v, res)
		}
	case *ast.ExprStmt:
		res = selfDepExprStrategy(stmt.X, res)
	default:
	}
	return res
}

func selfDepExprStrategy(expr ast.Expr, res []*selfDepCall) []*selfDepCall {
	switch expr := expr.(type) {
	case *ast.UnaryExpr:
		res = selfDepExprStrategy(expr.X, res)
	case *ast.StarExpr:
		res = selfDepExprStrategy(expr.X, res)
	case *ast.KeyValueExpr:
		res = selfDepExprStrategy(expr.Key, res)
		res = selfDepExprStrategy(expr.Value, res)
	case *ast.MapType:
		res = selfDepExprStrategy(expr.Key, res)
		res = selfDepExprStrategy(expr.Value, res)
	case *ast.CompositeLit:
		res = selfDepExprStrategy(expr.Type, res)
		for _, v := range expr.Elts {
			res = selfDepExprStrategy(v, res)
		}
	case *ast.BinaryExpr:
		res = selfDepExprStrategy(expr.X, res)
		res = selfDepExprStrategy(expr.Y, res)
	case *ast.FuncLit:
		return selfDepStmtStrategy(expr.Body, res)
	case *ast.SelectorExpr:
		return append(res, &selfDepCall{
			selectors: selfDepSelectorCrawler(expr.X, false),
			function:  expr.Sel.Name,
		})
	case *ast.CallExpr:
		res = selfDepExprStrategy(expr.Fun, res)
		for _, v := range expr.Args {
			res = selfDepExprStrategy(v, res)
		}
	// No Ops
	case *ast.BasicLit:
	case *ast.Ident:
	default:
	}
	return res
}

func selfDepSelectorCrawler(expr ast.Expr, isFunc bool) []string {
	switch expr := expr.(type) {
	case *ast.Ident:
		// Handle identifier nodes.
		return []string{expr.Name}
	case *ast.SelectorExpr:
		// Handle selector expression nodes.
		str := expr.Sel.Name
		if isFunc {
			str += "()"
		}
		return append(selfDepSelectorCrawler(expr.X, false), str)
	case *ast.IndexExpr:
		// Handle index expression nodes.
		return selfDepSelectorCrawler(expr.X, false)
	case *ast.CallExpr:
		return selfDepSelectorCrawler(expr.Fun, true)
	default:
		// Handle other types of nodes.
		return nil
	}
}

func receiverName(decl *ast.FuncDecl) string {
	if decl.Recv == nil {
		return ""
	}

	if len(decl.Recv.List) < 1 {
		return ""
	}

	return decl.Recv.List[0].Names[0].Name
}

func receiverType(decl *ast.FuncDecl) string {
	if decl.Recv == nil {
		return ""
	}

	if len(decl.Recv.List) < 1 {
		return ""
	}

	switch expr := decl.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.StarExpr:
		if res, ok := expr.X.(*ast.Ident); ok {
			return res.Name
		}
	}
	return ""
}
