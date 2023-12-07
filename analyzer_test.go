package canonicalheader_test

import (
	"net/http"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lasiar/canonicalheader"
)

const testValue = "hello_world"

func TestAnalyzer(t *testing.T) {
	t.Parallel()
	analysistest.RunWithSuggestedFixes(
		t,
		analysistest.TestData(),
		canonicalheader.Analyzer,
		"p",
	)
}

func BenchmarkCanonical(b *testing.B) {
	v := http.Header{
		"Canonical-Header": []string{testValue},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := v.Get("Canonical-Header")
		if s != testValue {
			b.Fatal()
		}
	}
}

func BenchmarkNonCanonical(b *testing.B) {
	v := http.Header{
		"Canonical-Header": []string{testValue},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := v.Get("CANONICAL-HEADER")
		if s != testValue {
			b.Fatal()
		}
	}
}
