package install

import (
	"fmt"
	"os"
	"path/filepath"
)

func RunUninstall() {
	dir := defaultInstallDir()

	for _, bin := range bins {
		p := filepath.Join(dir, bin)
		if err := os.Remove(p); err != nil {
			fmt.Printf("⚠️  Failed removing %s: %v\n", bin, err)
		} else {
			fmt.Printf("🗑️ Removed %s\n", bin)
		}
	}
}
