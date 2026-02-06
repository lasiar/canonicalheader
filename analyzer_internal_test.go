package canonicalheader

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeOfIdentMissing(t *testing.T) {
	t.Parallel()

	ident := &ast.Ident{Name: "x"}

	typ, ok := typeOfIdent(&types.Info{}, ident)

	require.False(t, ok)
	require.Nil(t, typ)
}

func TestTypeOfIdentPresent(t *testing.T) {
	t.Parallel()

	ident := &ast.Ident{Name: "x"}
	obj := types.NewVar(token.NoPos, nil, "x", types.Typ[types.Int])
	info := &types.Info{
		Uses: map[*ast.Ident]types.Object{
			ident: obj,
		},
	}

	typ, ok := typeOfIdent(info, ident)

	require.True(t, ok)
	require.Equal(t, types.Typ[types.Int], typ)
}
