package models

import (
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	//"fmt"
)

const (
	usersDbColumn = "authentication_users"
)

// LimitedPrintOutUser is the user that gets printed out to the client
type LimitedPrintOutUser struct {
	ID         bson.ObjectId
	Email      string
	Name       string
	IsAdmin    bool
	IsOperator bool
	IsActive   bool
	Title      string
	Rank       string
	Unit       string
	Image      string
}

// LimitedReadOut is used as a reduced function set of a user,
// that gets printed out to the client
func (u *User) LimitedReadOut() *LimitedPrintOutUser {
	return &LimitedPrintOutUser{
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		IsAdmin:    u.IsAdmin,
		IsOperator: u.IsOperator,
		IsActive:   u.IsActive,
		Title:      u.Title,
		Rank:       u.Rank,
		Unit:       u.Unit,
		Image:      u.Image,
	}
}

// User defines a single user object
type User struct {
	ID                   bson.ObjectId `bson:"_id,omitempty"`
	Email                string        `bson:"email"`
	EncryptedPassword    string        `bson:"encrypted_password"`
	Name                 string        `bson:"name"`
	Rank                 string        `bson:"rank"`
	Unit                 string        `bson:"unit"`
	Title                string        `bson:"title"`
	Role                 string        `bson:"role"`
	Image                string        `bson:"image"`
	IsActive             bool          `bson:"active"`
	IsAdmin              bool          `bson:"is_admin"`
	IsOperator           bool          `bson:"is_operator"`
	DomainID             bson.ObjectId `bson:"domain_id"`
	AuthenticationToken  string        `bson:"cur_authentication_token"`
	TrackingTokenLong    []string      `bson:"tracking_token_long"`
	TrackingTokenSession []string      `bson:"tracking_token_session"`
	TrackingTokenVisit   []string      `bson:"tracking_token_visit"`
	SignInCount          int           `bson:"sign_in_count"`
	CurrentSignInAt      time.Time     `bson:"current_sign_in_at"`
	LastSignInAt         time.Time     `bson:"last_sign_in_at"`
	CurrentSignInIP      string        `bson:"current_sign_in_ip"`
	LastSignInIP         string        `bson:"last_sign_in_ip"`
	HmacSigningKey       []byte        `bson:"hmac_secret"`
}

// Users define a list of users
type Users []User

// UserInvalidError defines an invalid user record
// Deprecated. Use InvalidError instead
type UserInvalidError struct {
	Message string
}

func (u *UserInvalidError) Error() string {
	return u.Message
}

// GetKeyForToken receives a key from cassandra based off a specific token
func GetKeyForToken(mongo *mgo.Database, rawToken string) ([]byte, error) {
	var user User
	err := PrepareQuery(mongo, usersDbColumn).Find(&bson.M{"cur_authentication_token": rawToken}).One(&user)
	if err != nil {
		return []byte{}, err
	}
	return user.HmacSigningKey, nil
}

// ValidateUserForDomain validates a user
func ValidateUserForDomain(mongo *mgo.Database, r *http.Request, accessToken string) error {
	return nil
	// token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
	// 	// if _, ok := token.Method.(jwt.GetSigningMethod("HS512")); !ok {
	// 	//     return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	// 	// }
	//
	// 	key, err := GetKeyForToken(mongo, token.Raw)
	// 	return key, err
	// })
	//
	// if err == nil && token.Valid {
	// 	return nil
	// }
	//
	// return err
}
