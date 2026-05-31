package hash

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const PasswordCost = 12

func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), PasswordCost)
	return string(b), err
}

func CheckPassword(plain, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// Token SHA-256 hashes a high-entropy random string (refresh token, invite token)
// safe to store directly in the DB — no need for bcrypt on random UUIDs
func Token(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}