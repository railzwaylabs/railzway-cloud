package nomad

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateJWTSecret generates a deterministic but unique JWT secret for each organization.
// This ensures each tenant has a unique secret while being reproducible from the master key.
func GenerateJWTSecret(masterKey string, orgID int64) string {
	if masterKey == "" {
		// Fallback to a default if master key is not configured
		masterKey = "default-insecure-key-change-in-production"
	}
	
	// Create deterministic hash: SHA256(masterKey + orgID)
	input := fmt.Sprintf("%s-%d", masterKey, orgID)
	hash := sha256.Sum256([]byte(input))
	
	// Return hex-encoded hash as JWT secret
	return hex.EncodeToString(hash[:])
}
