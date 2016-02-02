package sitrep

import (
	"time"

	"github.com/gocql/gocql"
)

// UsersSafeReturn represent safe user fields
type UsersSafeReturn struct {
	IsAdmin             bool
	IsAnalyzed          bool
	IsBanned            bool
	AccessValidTill     time.Time
	UserRank            string
	UserSelfDescription string
	TwitterName         string
	UserTitle           string
	UserUnit            string
	Email               string
	RealName            string
}

//MapUsersToSafe returns save user records
func MapUsersToSafe(iter *gocql.Iter) ([]UsersSafeReturn, error) {
	var array []UsersSafeReturn
	err := MapUsersByEmail(iter, func(t UsersByEmail) (bool, error) {
		array = append(array, UsersSafeReturn{
			IsAdmin:             t.IsAdmin,
			IsAnalyzed:          t.IsAnalyzed,
			IsBanned:            t.IsBanned,
			AccessValidTill:     t.AccessValidTill,
			UserRank:            t.UserRank,
			UserSelfDescription: t.UserSelfDescription,
			TwitterName:         t.TwitterName,
			UserTitle:           t.UserTitle,
			UserUnit:            t.UserUnit,
			Email:               t.Email,
			RealName:            t.RealName,
		})
		return true, nil
	})
	return array, err
}
