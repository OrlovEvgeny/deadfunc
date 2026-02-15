package deadfunc_test

import (
	"testing"

	"github.com/OrlovEvgeny/deadfunc"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, deadfunc.Analyzer, "example")
}

func TestAnalyzerWithExported(t *testing.T) {
	testdata := analysistest.TestData()
	// Enable exported function checking.
	if err := deadfunc.Analyzer.Flags.Set("exported", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = deadfunc.Analyzer.Flags.Set("exported", "false")
	}()
	analysistest.Run(t, testdata, deadfunc.Analyzer, "exported")
}
