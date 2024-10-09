package snaps

import (
	"fmt"
	"os"
)

// RemoveSnap deletes the snap file identified by the given filePath.
// It removes the file from the snap directory and confirms successful deletion.
func RemoveSnap(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove snap file: %v", err)
	}
	return nil
}
