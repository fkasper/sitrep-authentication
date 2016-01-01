package models

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gocql/gocql"
	"github.com/vatcinc/bio/schema"
	"golang.org/x/crypto/bcrypt"
	//"fmt"
)

// USERS is a reference to the users table in cassandra
var USERS = bio.UsersTableDef()

// LimitedPrintOutUser is the user that gets printed out to the client
type LimitedPrintOutUser struct {
	ID    gocql.UUID `json:"_id,omitempty"`
	Email string     `json:"email"`
	Name  string     `json:"name"`
}

// NewLimitedUser is used as a reduced function set of a user,
// that gets printed out to the client
func NewLimitedUser(user bio.Users) LimitedPrintOutUser {
	return LimitedPrintOutUser{
		ID:    user.Id,
		Email: user.Email,
		Name:  user.Name,
	}
}

// JWTResponseFormat is an RFC conform response for the OAUTH Standard
type JWTResponseFormat struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// NewJWTResponse prints out a new OAuth token and format
func NewJWTResponse(token string) *JWTResponseFormat {
	return &JWTResponseFormat{
		AccessToken: token,
		TokenType:   "bearer",
	}
}

// Users define a list of users
type Users []bio.Users

// UserInvalidError defines an invalid user record
// Deprecated. Use InvalidError instead
type UserInvalidError struct {
	Message string
}

func (u *UserInvalidError) Error() string {
	return u.Message
}

// GetKeyForToken receives a key from cassandra based off a specific token
func GetKeyForToken(cassandra *gocql.ClusterConfig, rawToken string) ([]byte, error) {
	var user bio.Users
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(USERS).
		Where(
		USERS.ACCESS_TOKEN.Eq(rawToken)).
		Into(
		USERS.To(&user)).
		FetchOne(session)
	//err := user.ByAttr(cassandra, "access_token", rawToken, `hmac_signing_key`)
	if err != nil {
		return []byte{}, err
	}
	if user.Email == "" {
		return []byte{}, &UserInvalidError{Message: "No Such User"}
	}
	if !user.IsActive {
		return []byte{}, &UserInvalidError{Message: "No Such User"}
	}
	return user.HmacSigningKey, nil
}

// ValidateUserForDomain validates a user
func ValidateUserForDomain(cassandra *gocql.ClusterConfig, r *http.Request, accessToken string) error {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		// if _, ok := token.Method.(jwt.GetSigningMethod("HS512")); !ok {
		//     return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		// }
		key, err := GetKeyForToken(cassandra, token.Raw)
		return key, err
	})

	if err == nil && token.Valid {
		return nil
	}

	return err
}

// SignInUser signs in a user using specified credentials
func SignInUser(cassandra *gocql.ClusterConfig, email string, password string, scope string) (interface{}, error) {
	var user bio.Users
	if email == "" || password == "" {
		return user, &UserInvalidError{Message: "No Password or Email"} //TODO: Error
	}
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(USERS).
		Where(
		USERS.EMAIL.Eq(email)).
		Into(
		USERS.To(&user)).
		FetchOne(session)

	if err != nil {
		return user, err
	}
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims["sub"] = user.Id
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	randKey := RandStringBytesMaskImprSrc(40)
	tkn, err := token.SignedString(randKey)
	if err != nil {
		return user, err
	}
	user.AccessToken = tkn
	user.HmacSigningKey = randKey

	if err := ctx.Upsert(USERS).
		SetString(USERS.ACCESS_TOKEN, tkn).
		SetBytes(USERS.HMAC_SIGNING_KEY, randKey).
		Where(
		USERS.EMAIL.Eq(email)).
		Exec(session); err != nil {
		return user, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(password)); err != nil {
		return user, err
	}
	return NewJWTResponse(user.AccessToken), nil
}
