package security

import "golang.org/x/crypto/bcrypt"

func toByte(value string) []byte {
	return []byte(value)
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword(toByte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hashedPass, givenPass string) error {
	return bcrypt.CompareHashAndPassword(toByte(hashedPass), toByte(givenPass))
}
