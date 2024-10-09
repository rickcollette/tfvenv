package snaps

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveSnap saves the provided Snap to a file with the given filePath.
// It marshals the Snap into JSON, encrypts the data, and writes it to the snap file.
func SaveSnap(filePath string, snap *Snap) error {
	// Convert snap data to JSON
	snapData, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("failed to marshal snap: %v", err)
	}

	// Encrypt the snap data
	encryptedData, err := Encrypt(snapData)
	if err != nil {
		return fmt.Errorf("failed to encrypt snap data: %v", err)
	}

	// Write the encrypted data to file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create snap file: %v", err)
	}
	defer file.Close()

	_, err = file.Write([]byte(encryptedData))
	if err != nil {
		return fmt.Errorf("failed to write snap file: %v", err)
	}

	return nil
}
