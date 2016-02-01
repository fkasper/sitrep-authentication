package utils

import (
	"time"

	"github.com/fkasper/sitrep-authentication/schema"
)

// APIUser represents safe to use user fields
type APIUser struct {
	UserEmail       string    `json:"email"`
	IsAdmin         bool      `json:"is_admin"`
	Rank            string    `json:"rank"`
	Title           string    `json:"title"`
	SelfDescription string    `json:"self_description"`
	TwitterAlias    string    `json:"twitter_alias"`
	IsAnalyzed      bool      `json:"has_tracking"`
	TrackingID      string    `json:"tracking_id"`
	LastLoggedIn    time.Time `json:"last_login"`
}

// MapUser polishes a user record for an API. Removes sensitive Data.
func MapUser(user *sitrep.UsersByEmail) *APIUser {
	if user == nil {
		return nil
	}
	return &APIUser{
		UserEmail:       user.Email,
		IsAdmin:         user.IsAdmin,
		Rank:            user.UserRank,
		Title:           user.UserTitle,
		SelfDescription: user.UserSelfDescription,
		TwitterAlias:    user.TwitterName,
		IsAnalyzed:      user.IsAnalyzed,
		TrackingID:      user.AnalyticsUserTrackingToken,
		LastLoggedIn:    user.LastLoggedIn,
	}
}
