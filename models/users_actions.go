package models

import (
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
)

// UsersTable is a reference to the users cassandra table
var UsersTable = sitrep.UsersByEmailTableDef()

// UsersJwtTable is a reference to the JWT Tokens table
var UsersJwtTable = sitrep.UsersByJwtTableDef()

// FindUserByEmail receives a user object from the database
func FindUserByEmail(cassandra *gocql.ClusterConfig, email string) (*sitrep.UsersByEmail, error) {
	var user sitrep.UsersByEmail
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(UsersTable).
		Where(
		UsersTable.EMAIL.Eq(email)).
		Into(
		UsersTable.To(&user)).
		FetchOne(session)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserSignIn verifies and authenticates a user from database
func UserSignIn(cassandra *gocql.ClusterConfig, email string, password string, scope string) (*sitrep.JWTResponse, error) {

	user, err := FindUserByEmail(cassandra, email)
	if err != nil {
		return nil, NewUserInvalidError()
	}
	if err := user.ValidatePassword(password); err != nil {
		return nil, NewUserInvalidError()
	}
	if user.IsBanned {
		return nil, NewUserInvalidError()
	}

	jwtToken, err := sitrep.NewJwtResponse(user.JwtEncryptionKey, user.Email)
	if err != nil {
		return nil, NewUserInvalidError()
	}
	jwtUser := &sitrep.UsersByJwt{
		EncryptionKey: user.JwtEncryptionKey,
		Jwt:           jwtToken.AccessToken,
		UserEmail:     user.Email,
		UserName:      user.RealName,
	}
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	if err := ctx.Store(UsersJwtTable.Bind(*jwtUser)).Exec(session); err != nil {
		return nil, err
	}

	return jwtToken, nil
}

// VerifyUserRequest verfies a request - as efficient as possible.
func VerifyUserRequest(cassandra *gocql.ClusterConfig, accessToken string) (*sitrep.UsersByEmail, error) {
	var jwt sitrep.UsersByJwt
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(UsersJwtTable).
		Where(
		UsersJwtTable.JWT.Eq(accessToken)).
		Into(
		UsersJwtTable.To(&jwt)).
		FetchOne(session)

	if err != nil {
		return nil, err
	}
	token, err := jwt.Verify()
	if err != nil {
		return nil, err
	}
	user, err := FindUserByEmail(cassandra, token.Claims["sub"].(string))
	if err != nil {
		return nil, err
	}
	if user.IsBanned {
		return nil, NewUserInvalidError()
	}
	return user, nil
}

// UserChangePassword changes a users password, if they match the previous one
func UserChangePassword(cassandra *gocql.ClusterConfig, user *sitrep.UsersByEmail, oldPasswd string, newPasswd string) (*map[string]string, error) {
	if err := user.ValidatePassword(oldPasswd); err != nil {
		return nil, NewUserInvalidError()
	}
	user.EncryptedPassword = newPasswd
	if err := user.HashCryptPassword(); err != nil {
		return nil, err
	}
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	if err := ctx.Upsert(UsersTable).
		SetString(UsersTable.ENCRYPTED_PASSWORD, user.EncryptedPassword).
		Where(
		UsersTable.EMAIL.Eq(user.Email)).
		Exec(session); err != nil {
		return nil, err
	}
	return &map[string]string{"status": "changed"}, nil
}

// UserInvalidError holds the error, when a user has no permission to access an exercise
type UserInvalidError struct {
	Message string
}

// Error prints the UserBannedError
func (u *UserInvalidError) Error() string {
	return u.Message
}

// NewUserInvalidError produces a new UserInvalidError
func NewUserInvalidError() *UserInvalidError {
	return &UserInvalidError{
		Message: "We were not able to log you in!",
	}
}
