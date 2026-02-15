// Package plugin provides golangci-lint module plugin integration for deadfunc.
//
// This registers deadfunc as a module plugin. To use it, add the following to
// your .golangci.yml:
//
//	linters:
//	  enable:
//	    - deadfunc
//
// And build a custom golangci-lint binary that imports this package.
package plugin

import (
	"github.com/OrlovEvgeny/deadfunc"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("deadfunc", New)
}

// Plugin implements register.LinterPlugin.
type Plugin struct{}

// New creates a new Plugin instance.
func New(_ any) (register.LinterPlugin, error) {
	return &Plugin{}, nil
}

// BuildAnalyzers returns the deadfunc analyzer.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{deadfunc.Analyzer}, nil
}

// GetLoadMode returns the load mode required by the analyzer.
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
