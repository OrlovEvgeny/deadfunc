// Package deadfunc implements a Go linter that detects functions declared
// but never used in production code (test files are excluded from both
// declaration and usage analysis).
package deadfunc

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Config holds linter configuration.
type Config struct {
	// IncludeExported checks exported functions too.
	// By default only unexported functions are checked, because exported
	// functions in library packages may be used by external consumers.
	IncludeExported bool

	// IncludeMethods checks methods (not just top-level functions).
	IncludeMethods bool

	// SkipGenerated skips files with "// Code generated" header.
	SkipGenerated bool
}

var defaultConfig = Config{
	IncludeExported: false,
	IncludeMethods:  true,
	SkipGenerated:   true,
}

var includeExported bool
var includeMethods bool
var skipGenerated bool

// Analyzer is the deadfunc analysis.Analyzer.
var Analyzer = &analysis.Analyzer{
	Name:     "deadfunc",
	Doc:      "reports functions that are declared but never called/referenced in production code (tests excluded)",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&includeExported, "exported", defaultConfig.IncludeExported,
		"also check exported functions (risky for library packages)")
	Analyzer.Flags.BoolVar(&includeMethods, "methods", defaultConfig.IncludeMethods,
		"also check methods, not just top-level functions")
	Analyzer.Flags.BoolVar(&skipGenerated, "skip-generated", defaultConfig.SkipGenerated,
		"skip files with '// Code generated' header")
}

// funcID uniquely identifies a function or method within a package.
type funcID struct {
	// For functions: just the name. For methods: "TypeName.MethodName".
	name string
	pos  token.Pos
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Step 1: Determine which files are test files.
	testFiles := make(map[string]bool)
	for _, f := range pass.Files {
		filename := pass.Fset.File(f.Pos()).Name()
		if strings.HasSuffix(filename, "_test.go") {
			testFiles[filename] = true
		}
	}

	// Step 2: Collect all function declarations from non-test files.
	declared := make(map[types.Object]funcID)  // obj -> funcID
	objByName := make(map[string]types.Object) // for quick lookup

	for _, f := range pass.Files {
		filename := pass.Fset.File(f.Pos()).Name()
		if testFiles[filename] {
			continue
		}
		if skipGenerated && isGenerated(f) {
			continue
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Skip init() and main() — they are entry points.
			if fn.Name.Name == "init" || fn.Name.Name == "main" {
				continue
			}

			// Skip test helpers that somehow ended up in non-test files
			// (e.g. TestMain in main package).
			if strings.HasPrefix(fn.Name.Name, "Test") ||
				strings.HasPrefix(fn.Name.Name, "Bench") ||
				strings.HasPrefix(fn.Name.Name, "Fuzz") ||
				strings.HasPrefix(fn.Name.Name, "Example") {
				continue
			}

			isMethod := fn.Recv != nil

			// Should we check methods?
			if isMethod && !includeMethods {
				continue
			}

			// Should we check exported functions?
			if !includeExported && fn.Name.IsExported() {
				continue
			}

			obj := pass.TypesInfo.Defs[fn.Name]
			if obj == nil {
				continue
			}

			// Skip methods that satisfy known interfaces (likely required).
			if isMethod {
				if isInterfaceImpl(pass, obj) {
					continue
				}
			}

			// Build the ID.
			name := funcObjName(fn, obj)

			declared[obj] = funcID{
				name: name,
				pos:  fn.Name.Pos(),
			}
			objByName[name] = obj
		}
	}

	if len(declared) == 0 {
		return nil, nil
	}

	// Step 3: Walk all non-test files and collect every reference to declared
	// functions. We track usage via types.Object identity.
	used := make(map[types.Object]bool)

	nodeFilter := []ast.Node{
		(*ast.Ident)(nil),
		(*ast.SelectorExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		// Determine the file this node is in.
		pos := pass.Fset.Position(n.Pos())
		if testFiles[pos.Filename] {
			return // ignore test-file usage
		}

		var obj types.Object

		switch x := n.(type) {
		case *ast.Ident:
			obj = pass.TypesInfo.Uses[x]
		case *ast.SelectorExpr:
			sel, ok := pass.TypesInfo.Selections[x]
			if ok {
				obj = sel.Obj()
			} else {
				// Qualified ident (pkg.Func).
				obj = pass.TypesInfo.Uses[x.Sel]
			}
		}

		if obj == nil {
			return
		}

		if _, exists := declared[obj]; exists {
			used[obj] = true
		}
	})

	// Step 4: Report unused functions.
	for obj, fid := range declared {
		if used[obj] {
			continue
		}
		pass.Reportf(fid.pos, "function %q is declared but never used in production code", fid.name)
	}

	return nil, nil
}

// funcObjName returns a human-readable name for the function.
func funcObjName(fn *ast.FuncDecl, obj types.Object) string {
	if fn.Recv == nil {
		return fn.Name.Name
	}
	// Method — prepend receiver type name.
	recv := fn.Recv.List[0].Type
	return fmt.Sprintf("%s.%s", receiverTypeName(recv), fn.Name.Name)
}

// receiverTypeName extracts the type name from a receiver field.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr: // generic receiver T[P]
		return receiverTypeName(t.X)
	case *ast.IndexListExpr: // generic receiver T[P1, P2]
		return receiverTypeName(t.X)
	default:
		return "?"
	}
}

// isGenerated checks if a file has the "// Code generated" marker.
func isGenerated(f *ast.File) bool {
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "Code generated") &&
				strings.Contains(c.Text, "DO NOT EDIT") {
				return true
			}
		}
	}
	return false
}

// isInterfaceImpl checks if the given method object satisfies any interface
// visible in the current package. This is a heuristic — it checks if any
// interface type in scope has a method with the same name and signature.
func isInterfaceImpl(pass *analysis.Pass, obj types.Object) bool {
	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}

	sig := fn.Type().(*types.Signature)
	recv := sig.Recv()
	if recv == nil {
		return false
	}

	// Get the named type of the receiver.
	recvType := recv.Type()
	// Dereference pointer receiver.
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}

	methodName := fn.Name()

	// Check all interfaces in the package scope and imported scopes.
	scope := pass.Pkg.Scope()
	for _, name := range scope.Names() {
		scopeObj := scope.Lookup(name)
		named, ok := scopeObj.Type().(*types.Named)
		if !ok {
			continue
		}
		iface, ok := named.Underlying().(*types.Interface)
		if !ok {
			continue
		}
		for i := 0; i < iface.NumMethods(); i++ {
			m := iface.Method(i)
			if m.Name() == methodName {
				// Check if the receiver type implements this interface.
				if types.Implements(recvType, iface) ||
					types.Implements(types.NewPointer(recvType), iface) {
					return true
				}
			}
		}
	}

	// Also check well-known stdlib interfaces: error, fmt.Stringer, etc.
	// by checking common method names.
	wellKnownInterfaceMethods := map[string]bool{
		"Error":           true, // error
		"String":          true, // fmt.Stringer
		"MarshalJSON":     true, // json.Marshaler
		"UnmarshalJSON":   true, // json.Unmarshaler
		"MarshalText":     true,
		"UnmarshalText":   true,
		"MarshalBinary":   true,
		"UnmarshalBinary": true,
		"Read":            true, // io.Reader
		"Write":           true, // io.Writer
		"Close":           true, // io.Closer
		"Len":             true, // sort.Interface
		"Less":            true,
		"Swap":            true,
		"ServeHTTP":       true, // http.Handler
		"Format":          true, // fmt.Formatter
		"Scan":            true, // fmt.Scanner
		"Value":           true, // driver.Valuer
		"GoString":        true, // fmt.GoStringer
		"WriteHeader":     true, // http.ResponseWriter
	}

	return wellKnownInterfaceMethods[methodName]
}
