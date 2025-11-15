package pkg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithUserIDAndGetUserID(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-123")

	require.Equal(t, "user-123", GetUserID(ctx))
}

func TestGetUserIDPanicsWhenMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when user id missing")
		}
	}()

	_ = GetUserID(context.Background())
}
