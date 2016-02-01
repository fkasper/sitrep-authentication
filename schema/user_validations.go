package sitrep

import "golang.org/x/crypto/bcrypt"

// ValidatePassword validates a password against the one, received from the Database
func (u *UsersByEmail) ValidatePassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password)); err != nil {
		return err
	}
	return nil
}

//HashCryptPassword encrypts the current user password
func (u *UsersByEmail) HashCryptPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.EncryptedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.EncryptedPassword = string(hashedPassword)
	return nil
}
