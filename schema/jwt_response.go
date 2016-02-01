package sitrep

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTResponse holds the structure for an OAuth2.0 Bearer Token
type JWTResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// NewJwtResponse creates a new JWT response object
func NewJwtResponse(key string, subj string) (*JWTResponse, error) {
	accessToken, err := generateToken([]byte(key), subj)
	if err != nil {
		return nil, err
	}
	return &JWTResponse{
		AccessToken: accessToken,
		Scope:       "exercise",
		TokenType:   "bearer",
	}, nil
}

func generateToken(key []byte, subj string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims["sub"] = subj
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	accessToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

// Verify validates a JWT token against its validity for a given key. It returns the token, if valid
func (j *UsersByJwt) Verify() (*jwt.Token, error) {
	token, err := jwt.Parse(j.Jwt, func(token *jwt.Token) (interface{}, error) {
		// if _, ok := token.Method.(jwt.GetSigningMethod("HS512")); !ok {
		//     return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		// }
		return []byte(j.EncryptionKey), nil
	})

	if err == nil && token.Valid {
		return token, nil
	}
	return nil, err
}
