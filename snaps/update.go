package snaps

import "fmt"

// UpdateSnap updates the snap information in the existing snap file identified by filePath.
// It first checks if the snap file exists and then saves the updated Snap data.
func UpdateSnap(filePath string, updatedSnap *Snap) error {
	// First, check if the file exists
	if _, err := GetSnap(filePath); err != nil {
		return fmt.Errorf("cannot update snap: %v", err)
	}
	// If it exists, save the updated information
	return SaveSnap(filePath, updatedSnap)
}
