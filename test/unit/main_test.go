package unit_test

import (
	"testing"
)

func TestBasicSetup(t *testing.T) {
	t.Run("should pass basic test", func(t *testing.T) {
		got := 1 + 1
		want := 2

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
}

func TestVersionFormat(t *testing.T) {
	t.Run("version should follow semver", func(t *testing.T) {
		version := "0.1.0-dev"

		if version == "" {
			t.Error("version should not be empty")
		}

		if len(version) < 5 {
			t.Error("version should be at least 5 characters (e.g., 0.1.0)")
		}
	})
}
