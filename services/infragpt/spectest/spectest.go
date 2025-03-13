package spectest

import "testing"

func Ensure(t *testing.T) {
	t.Run("AskForAccess", func(t *testing.T) {
		t.Run("fails if", func(t *testing.T) {
			t.Run("UserID is empty", func(t *testing.T) {
				t.Skip("skipping")
			})
		})
	})
}
