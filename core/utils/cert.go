package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
)

// ComputeCertSHA256 parses a PEM-encoded certificate and returns
// the lowercase hex SHA256 of the DER-encoded leaf certificate.
func ComputeCertSHA256(pemText string) (string, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}
	if !strings.HasSuffix(block.Type, "CERTIFICATE") {
		return "", fmt.Errorf("PEM block type %q is not a certificate", block.Type)
	}
	sum := sha256.Sum256(block.Bytes)
	return hex.EncodeToString(sum[:]), nil
}
