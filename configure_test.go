package canonicalheader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSetSetAppends(t *testing.T) {
	t.Parallel()

	var s stringSet

	require.NoError(t, s.Set("a, b"))
	require.NoError(t, s.Set("c"))

	require.Equal(t, stringSet{"a", "b", "c"}, s)
}
