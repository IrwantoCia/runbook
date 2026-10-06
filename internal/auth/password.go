package auth

import "golang.org/x/crypto/bcrypt"

var dummyHash, _ = HashPassword("dummy-password-for-timing")

func DummyHash() string { return dummyHash }

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
