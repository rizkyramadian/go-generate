package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"strings"

	"github.com/rizkyramadian/go-generate/extractor"
	"github.com/rizkyramadian/go-generate/templates"

	"github.com/spf13/pflag"
)

const version = "0.0.1"

var checkVar bool

func init() {
	pflag.BoolVarP(&checkVar, "version", "v", false, "Checks Version")
}

func main() {
	pflag.Parse()

	if checkVar {
		fmt.Printf("Version %v \n", version)
		return
	}

	// Incase Args is not proper
	if pflag.NArg() != 3 {
		fmt.Println("No arguments given, please use the following format: go-generate <package dir> <package dir destination> <pkg name dest> [flags...]")
		fmt.Println("usage: go-generate <package dir source> <package dir destination> <pkg name dest> [flags...]")
		fmt.Println("use -h | --help for help")
		return
	}

	// Prepare variables from args
	dir := pflag.Arg(0)
	destDir := pflag.Arg(1)
	packageName := pflag.Arg(2)
	newAlias := string(packageName[0])

	// New Token File set for each new package
	fset := token.NewFileSet()
	destfset := token.NewFileSet()

	// Setup AST Parser
	pkgs, err := parser.ParseDir(fset, dir, func(x fs.FileInfo) bool {
		return !strings.Contains(x.Name(), "_test")
	}, 0)
	if err != nil {
		fmt.Println("Failed to open dir:", dir)
		fmt.Println(err)
		return
	}

	// Setup AST parser for destination directory
	dest, err := parser.ParseDir(destfset, destDir, func(x fs.FileInfo) bool {
		return !strings.Contains(x.Name(), "_test")
	}, 0)
	if err != nil {
		fmt.Println("Failed to open dir:", dir)
		fmt.Println(err)
		return
	}

	// Extract functions visitor
	fn := extractor.NewFunctionVisitor()

	// Run function visitor on the source package
	for _, v := range pkgs {
		ast.Walk(fn, v)
	}

	// Scan Destination Interface
	destScanner := extractor.NewDestInterfaceScanner()
	defPkg := ""
	for k, v := range dest {
		defPkg = k
		ast.Walk(destScanner, v)
	}

	// Create Destination Directory
	newPath := destDir + "/" + packageName
	err = os.Mkdir(newPath, os.ModePerm)
	if err != nil {
		fmt.Println("Error ", err)
	}

	// Create Init File (based on template)
	initWriter, err := os.Create(newPath + "/init.go")
	if err != nil {
		fmt.Println("Error Init ", err)
	}
	initWriter.Write([]byte(templates.InitFile(packageName, destScanner.InterfaceName, defPkg)))
	defer initWriter.Close()

	// Create Func Go file implementation
	funcFile, err := os.Create(newPath + "/func.go")
	if err != nil {
		fmt.Print(err)
	}
	defer funcFile.Close()

	// Extract needed functions
	funcs := make([]ast.Decl, 0)
	for _, v := range destScanner.Functions {
		if f, ok := fn.Functions[v]; ok {
			// Rename funcs selector to new
			ncw := extractor.NewNameChangerWalker(newAlias, packageName+"Impl", receiverName(f), receiverType(f))
			ast.Walk(ncw, f)
			funcs = append(funcs, f.(ast.Decl))
		}
	}

	// Prepare New File
	newFileAst := &ast.File{
		Doc:     nil,
		Package: 0,
		Name: &ast.Ident{
			Name: packageName,
		},
		Decls: funcs,
	}

	// Print AST to new function file
	printer.Fprint(funcFile, token.NewFileSet(), newFileAst)

}

func receiverName(n ast.Node) string {
	decl, ok := n.(*ast.FuncDecl)
	if !ok {
		return ""
	}

	if decl.Recv == nil {
		return ""
	}

	if len(decl.Recv.List) < 1 {
		return ""
	}

	return decl.Recv.List[0].Names[0].Name
}

func receiverType(n ast.Node) string {
	decl, ok := n.(*ast.FuncDecl)
	if !ok {
		return ""
	}

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
