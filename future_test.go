package future

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletable(t *testing.T) {
	c := NewCompleatableFuture[string]()
	require.NoError(t, c.Complete("hello"))
	v, err := c.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "hello", v)
}
