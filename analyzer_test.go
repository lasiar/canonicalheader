package canonicalheader_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lasiar/canonicalheader"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()
	analysistest.RunWithSuggestedFixes(
		t,
		analysistest.TestData(),
		canonicalheader.Analyzer,
		"p",
	)
}
