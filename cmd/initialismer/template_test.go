package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateNameForTarget(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		target   generateTarget
		wantName string
		wantErr  bool
	}{
		{
			name:     "test",
			target:   generateTest,
			wantName: templateTest,
		},
		{
			name:     "test-golden",
			target:   generateTestGolden,
			wantName: templateGolden,
		},
		{
			name:     "mapping",
			target:   generateMapping,
			wantName: templateMapping,
		},
		{
			name:    "unknown",
			target:  generateUnknown,
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := templateNameForTarget(tt.target)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, name)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantName, name)
		})
	}
}
