package auth

import "golang.org/x/crypto/bcrypt"

type PasswordService struct {
	cost int
}

func NewPasswordService(cost int) PasswordService {
	return PasswordService{cost: cost}
}

func (p PasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	return string(hash), err
}

func (p PasswordService) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
