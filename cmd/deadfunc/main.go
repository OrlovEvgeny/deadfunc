// Command deadfunc runs the deadfunc analyzer.
//
// Usage:
//
//	deadfunc [-exported] [-methods=false] [-skip-generated=false] ./...
package main

import (
	"github.com/OrlovEvgeny/deadfunc"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(deadfunc.Analyzer)
}
