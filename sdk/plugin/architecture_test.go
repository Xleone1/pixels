package plugin

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestSDKDoesNotDependOnInternalPackages protects the plugin compilation boundary.
func TestSDKDoesNotDependOnInternalPackages(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	command := exec.Command("go", "list", "-deps", "./sdk/...")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatalf("list sdk dependencies: %v", err)
	}
	for _, dependency := range strings.Fields(string(output)) {
		if strings.HasPrefix(dependency, "github.com/niflaot/pixels/internal/") {
			t.Fatalf("SDK depends on private package %q", dependency)
		}
	}
}
