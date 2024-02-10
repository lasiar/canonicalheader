package funcliner_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lasiar/funcliner"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()
	analysistest.RunWithSuggestedFixes(
		t,
		analysistest.TestData(),
		funcliner.Analyzer,
		"p",
	)
}
