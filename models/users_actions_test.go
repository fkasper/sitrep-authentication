package models_test

import (
	"testing"
	"time"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
	"github.com/relops/cqlc/cqlc"
)

// UsersTable is a reference to the users cassandra table
var UsersTable = sitrep.UsersByEmailTableDef()

// UsersJwtTable ;)
var UsersJwtTable = sitrep.UsersByJwtTableDef()

func dbConn() *gocql.ClusterConfig {
	db := gocql.NewCluster("127.0.0.1:9042")
	db.Keyspace = "sitrep"
	return db
}
func mockDb() (*gocql.Session, *cqlc.Context) {
	conn, context, err := models.WithSession(dbConn())
	if err != nil {
		panic(err)
	}
	return conn, context
}

func mockUser() *sitrep.UsersByEmail {
	return &sitrep.UsersByEmail{
		Email:             "someguy@somedomain.com",
		IsAdmin:           true,
		IsAnalyzed:        true,
		IsBanned:          false,
		IsExpiring:        false,
		JwtEncryptionKey:  "somekey",
		UserRank:          "LTC",
		EncryptedPassword: "$2a$04$5/1FR1jhqkxr92WvuXsa8ep46PoSBbuuwM6SLuqptqG5j4f3lQyTS",
		LastLoggedIn:      time.Now(),
	}
}

func mockJwtUser(jwt string) *sitrep.UsersByJwt {
	return &sitrep.UsersByJwt{
		EncryptionKey: "somekey",
		Jwt:           jwt,
		UserEmail:     "someguy@somedomain.com",
		UserName:      "Guy, Some",
	}
}
func initUser(user *sitrep.UsersByEmail) {
	var usr *sitrep.UsersByEmail
	if user != nil {
		usr = user
	} else {
		usr = mockUser()
	}
	session, ctx := mockDb()
	defer session.Close()
	err := ctx.Store(UsersTable.Bind(*usr)).Exec(session)
	if err != nil {
		panic(err)
	}
}

func initJwtUser(user *sitrep.UsersByJwt, jwt string) {
	var usr *sitrep.UsersByJwt
	if user != nil {
		usr = user
	} else {
		usr = mockJwtUser(jwt)
	}
	session, ctx := mockDb()
	defer session.Close()
	err := ctx.Store(UsersJwtTable.Bind(*usr)).Exec(session)
	if err != nil {
		panic(err)
	}
}

func TestUser_Authentication_WithoutData(t *testing.T) {
	initUser(nil)
	_, err := models.UserSignIn(dbConn(), "", "", "")
	if err == nil {
		t.Fatalf("User was signed in without an email oO")
	}
}

func TestUser_Authentication_WithInCorrectPassword(t *testing.T) {
	initUser(nil)
	_, err := models.UserSignIn(dbConn(), "someguy@somedomain.com", "test1235", "password")
	if err == nil {
		t.Fatalf("Incorrect password was accepted!")
	}
}

func TestUser_Authentication_WithCorrectPassword(t *testing.T) {
	initUser(nil)
	_, err := models.UserSignIn(dbConn(), "someguy@somedomain.com", "test1234", "password")
	if err != nil {
		t.Fatalf("Correct password was not accepted! %v", err.Error())
	}
}

func TestUser_IsBanned(t *testing.T) {
	user := mockUser()
	user.IsBanned = true
	initUser(user)
	_, err := models.UserSignIn(dbConn(), "someguy@somedomain.com", "test1234", "password")
	if err == nil {
		t.Fatalf("Banned User was allowed into the system")
	}
	if err.Error() != "We were not able to log you in!" {
		t.Fatalf("Wrong message was printed")
	}
}

func TestUser_AccessTokenValid(t *testing.T) {
	//VerifyUserRequest
	initUser(nil)
	c := dbConn()
	user, err := models.UserSignIn(c, "someguy@somedomain.com", "test1234", "password")
	if err != nil {
		t.Fatalf("Sign in failed unexpectedly")
	}
	initJwtUser(nil, user.AccessToken)

	if _, err := models.VerifyUserRequest(c, user.AccessToken); err != nil {
		t.Fatalf("Access token verification failed")
	}
}

func TestUser_AccessTokenInValid(t *testing.T) {
	//VerifyUserRequest
	initUser(nil)
	fuser := mockJwtUser("1234")
	c := dbConn()
	user, err := models.UserSignIn(c, "someguy@somedomain.com", "test1234", "password")
	if err != nil {
		t.Fatalf("Sign in failed unexpectedly")
	}
	initJwtUser(fuser, user.AccessToken)

	if _, err := models.VerifyUserRequest(c, "1234"); err == nil {
		t.Fatalf("Access token accidientially Verified. Should be false")
	}
}

func TestUser_Change_Passwd_ValidCurrent(t *testing.T) {
	//UserChangePassword
	initUser(nil)
	c := dbConn()
	req, err := models.UserSignIn(c, "someguy@somedomain.com", "test1234", "password")
	if err != nil {
		t.Fatalf("login failed unexpectedly")
		return
	}
	u, err := models.VerifyUserRequest(c, req.AccessToken)
	if _, err := models.UserChangePassword(c, u, "test1234", "test12345"); err != nil {
		t.Fatalf("password change failed unexpectedly")
		return
	}

	if _, err := models.UserSignIn(c, "someguy@somedomain.com", "test12345", "password"); err != nil {
		t.Fatalf("second login failed unexpectedly")
		return
	}

}

func TestUser_Change_Passwd_In_ValidCurrent(t *testing.T) {
	//UserChangePassword
	initUser(nil)
	c := dbConn()
	req, err := models.UserSignIn(c, "someguy@somedomain.com", "test1234", "password")
	if err != nil {
		t.Fatalf("login failed unexpectedly")
		return
	}
	u, err := models.VerifyUserRequest(c, req.AccessToken)
	if _, err := models.UserChangePassword(c, u, "test12355", "test12345"); err == nil {
		t.Fatalf("password change was unexpectedly successful")
		return
	}

}
func TestUser_FetchAll(t *testing.T) {
	//UserChangePassword
	initUser(nil)
	c := dbConn()
	users, err := models.FetchAllUsers(c)
	if err != nil {
		t.Fatalf("fetch failed unexpectedly")
		return
	}

	if len(users) < 1 {
		t.Fatalf("could not fetch users. lol")
		return
	}
}
