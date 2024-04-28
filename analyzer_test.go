package canonicalheader_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lasiar/canonicalheader"
)

const testValue = "hello_world"

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := [...]string{
		"alias",
		"common",
		"embedded",
		"global",
		"struct",
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			t.Parallel()

			analysistest.RunWithSuggestedFixes(
				t,
				analysistest.TestData(),
				canonicalheader.Analyzer,
				tt,
			)
		})
	}

	t.Run("are_test_cases_complete", func(t *testing.T) {
		t.Parallel()

		dirs, err := os.ReadDir(filepath.Join(analysistest.TestData(), "src"))
		require.NoError(t, err)
		require.Len(t, testCases, len(dirs))

		for i, dir := range dirs {
			require.Equal(t, dir.Name(), testCases[i])
		}
	})
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
