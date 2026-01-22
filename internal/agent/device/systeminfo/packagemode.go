package systeminfo

import (
	"os/exec"
)

// DetectPackageMode determines if the system is running in package-mode (traditional package management)
// or image-mode (bootc-based OS management).
//
// Package-mode is detected by the absence of the bootc binary in the system PATH.
// This detection runs once at agent startup and the result is cached for the agent's lifetime.
//
// Returns:
//   - true: Package-mode (OS managed via dnf/apt, bootc not present)
//   - false: Image-mode (OS managed via bootc, bootc binary found)
func DetectPackageMode() bool {
	// Check if bootc binary exists in PATH
	_, err := exec.LookPath("bootc")

	// If bootc is not found, we're in package-mode
	return err != nil
}
