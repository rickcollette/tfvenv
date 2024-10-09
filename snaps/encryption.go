package snaps

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "fmt"
    "io"
    "os"
)

// getSnapKey retrieves the SNAP_KEY from environment variables and ensures it is the correct length.
func getSnapKey() ([]byte, error) {
    snapKey := os.Getenv("SNAP_KEY")
    if len(snapKey) != 32 {
        return nil, fmt.Errorf("invalid SNAP_KEY, must be 32 bytes long for AES-256 encryption")
    }
    return []byte(snapKey), nil
}

// Encrypt encrypts the data using AES-CTR and returns the encrypted data along with an HMAC for integrity.
func Encrypt(data []byte) (string, error) {
    encryptionKey, err := getSnapKey()
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(encryptionKey)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, aes.BlockSize)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    stream := cipher.NewCTR(block, nonce)
    ciphertext := make([]byte, len(data))
    stream.XORKeyStream(ciphertext, data)

    // Compute HMAC for integrity
    h := hmac.New(sha256.New, encryptionKey)
    h.Write(ciphertext)
    mac := h.Sum(nil)

    // Combine nonce, ciphertext, and MAC
    combined := append(nonce, ciphertext...)
    combined = append(combined, mac...)
    return base64.StdEncoding.EncodeToString(combined), nil
}


// Decrypt decrypts the encrypted data and validates the HMAC for integrity.
func Decrypt(encrypted []byte) ([]byte, error) {
	encryptionKey, err := getSnapKey()
	if err != nil {
		return nil, err
	}

	nonceSize := aes.BlockSize
	macSize := sha256.Size

	// Ensure data length is sufficient
	if len(encrypted) < nonceSize+macSize {
		return nil, errors.New("invalid encrypted data length")
	}

	// Extract nonce, ciphertext, and MAC
	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize : len(encrypted)-macSize]
	mac := encrypted[len(encrypted)-macSize:]

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Validate HMAC
	h := hmac.New(sha256.New, encryptionKey)
	h.Write(ciphertext)
	expectedMac := h.Sum(nil)
	if !hmac.Equal(mac, expectedMac) {
		return nil, errors.New("invalid MAC")
	}

	// Decrypt using AES-CTR
	stream := cipher.NewCTR(block, nonce)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)
	return plaintext, nil
}