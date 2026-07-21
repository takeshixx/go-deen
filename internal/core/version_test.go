package core

import "testing"

func TestVersion(t *testing.T) {
	originalVersion, originalBranch := version, branch
	t.Cleanup(func() {
		version, branch = originalVersion, originalBranch
	})

	t.Run("development fallback", func(t *testing.T) {
		version, branch = "", ""
		if got := Version(); got != "dev" {
			t.Fatalf("Version() = %q, want %q", got, "dev")
		}
		if got := BuildVersion(); got != "dev" {
			t.Fatalf("BuildVersion() = %q, want %q", got, "dev")
		}
	})

	t.Run("stable release", func(t *testing.T) {
		version, branch = "v3.4.0", ""
		if got := BuildVersion(); got != "v3.4.0" {
			t.Fatalf("BuildVersion() = %q, want %q", got, "v3.4.0")
		}
	})

	t.Run("development branch", func(t *testing.T) {
		version, branch = "v3.4.0-beta", "master"
		if got := BuildVersion(); got != "v3.4.0-beta-master" {
			t.Fatalf("BuildVersion() = %q, want %q", got, "v3.4.0-beta-master")
		}
	})
}
