package snaps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// GetSnap retrieves a Snap by its filePath.
// It reads the encrypted snap file, decrypts it, and unmarshals the JSON data into a Snap struct.
func GetSnap(filePath string) (*Snap, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open snap file: %v", err)
	}
	defer file.Close()

	// Read the encrypted data
	encryptedData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read snap file: %v", err)
	}

	// Decode the base64 string to bytes
	decodedData, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode snap data: %v", err)
	}

	// Decrypt the data
	decryptedData, err := Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt snap data: %v", err)
	}

	// Decode the JSON into Snap struct
	var snap Snap
	err = json.Unmarshal(decryptedData, &snap)
	if err != nil {
		return nil, fmt.Errorf("failed to decode snap data: %v", err)
	}

	return &snap, nil
}
