package analyer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	analyzer "github.com/Lasiar/canonicalHeader"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()
	analysistest.RunWithSuggestedFixes(
		t,
		analysistest.TestData(),
		analyzer.Analyzer,
		"p",
	)
}
