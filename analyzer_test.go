package canonicalheader_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lasiar/canonicalheader"
)

const testValue = "hello_world"

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		pkg            string
		customAnalyzer func(t *testing.T) *analysis.Analyzer
	}{
		{pkg: "alias"},
		{pkg: "assigned"},
		{pkg: "common"},
		{pkg: "const"},
		{pkg: "embedded"},
		{
			pkg: "exclusions",
			customAnalyzer: func(t *testing.T) *analysis.Analyzer {
				t.Helper()

				a := canonicalheader.New()
				err := a.Flags.Set("exclusions", "exclusioN")
				require.NoError(t, err)
				return a
			},
		},
		{pkg: "global"},
		{pkg: "initialism"},
		{pkg: "struct"},
		{pkg: "underlying"},
	}

	for _, tt := range testCases {
		t.Run(tt.pkg, func(t *testing.T) {
			t.Parallel()

			var a *analysis.Analyzer
			if tt.customAnalyzer != nil {
				a = tt.customAnalyzer(t)
			} else {
				a = canonicalheader.New()
			}

			analysistest.RunWithSuggestedFixes(
				t,
				analysistest.TestData(),
				a,
				tt.pkg,
			)
		})
	}

	t.Run("are_test_cases_complete", func(t *testing.T) {
		t.Parallel()

		want, err := os.ReadDir(filepath.Join(analysistest.TestData(), "src"))
		require.NoError(t, err)
		require.Len(t, testCases, len(want))

		got := [len(testCases)]string{}
		for i, testCase := range testCases {
			got[i] = testCase.pkg
		}

		require.EqualValues(
			t,
			transform(want, func(d os.DirEntry) string {
				return d.Name()
			}),
			got,
		)
	})
}

func transform[S ~[]E, E any, T any](sl S, f func(E) T) []T {
	out := make([]T, len(sl))
	for i, t := range sl {
		out[i] = f(t)
	}

	return out
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
