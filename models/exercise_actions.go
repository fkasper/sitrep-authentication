package models

import (
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
	"github.com/imdario/mergo"
)

// ExerciseByIdentifierTable is a reference to the users cassandra table
var ExerciseByIdentifierTable = sitrep.ExerciseByIdentifierTableDef()

// UsersInExerciseTable is a reference to the users cassandra table
var UsersInExerciseTable = sitrep.CreateUsersInExerciseTableDef()

// ExercisePermissionsLevelTable is a reference to the users cassandra table
var ExercisePermissionsLevelTable = sitrep.ExercisePermissionsLevelTableDef()

// SettingsByExerciseIdentifierTable is a reference to the SettingsByExerciseIdentifier cassandra table
var SettingsByExerciseIdentifierTable = sitrep.SettingsByExerciseIdentifierTableDef()

// FindExerciseByID receives an exercise from the database
func FindExerciseByID(cassandra *gocql.ClusterConfig, id gocql.UUID) (*sitrep.ExerciseByIdentifier, error) {
	var exercisesMap sitrep.ExerciseByIdentifier
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(ExerciseByIdentifierTable).
		Where(
		ExerciseByIdentifierTable.ID.Eq(id)).
		Into(
		ExerciseByIdentifierTable.To(&exercisesMap)).
		FetchOne(session)

	if err != nil {
		return nil, err
	}
	return &exercisesMap, nil
}

// FindExercisesForUser receives exercises for a user
func FindExercisesForUser(cassandra *gocql.ClusterConfig, user *sitrep.UsersByEmail) (*sitrep.CreateUsersInExercise, error) {
	if user == nil {
		return nil, NewUserInvalidError()
	}
	var exercisesMap sitrep.CreateUsersInExercise
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(UsersInExerciseTable).
		Where(
		UsersInExerciseTable.EMAIL.Eq(user.Email)).
		Into(
		UsersInExerciseTable.To(&exercisesMap)).
		FetchOne(session)

	if err != nil {
		return nil, err
	}
	return &exercisesMap, nil
}

// FindExercisePermissionsForUser receives a new permission model from Cassandra
// Exercise Permissions are a link between users and exercises
// They also add permissions for users
func FindExercisePermissionsForUser(cassandra *gocql.ClusterConfig, user *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier) (*sitrep.ExercisePermissionsLevel, error) {
	if user == nil {
		return nil, NewUserInvalidError() //TBD
	}
	if exercise == nil {
		return nil, NewUserInvalidError() //TBD
	}
	var permissionsMap sitrep.ExercisePermissionsLevel
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	_, err := ctx.Select().
		From(ExercisePermissionsLevelTable).
		Where(
		ExercisePermissionsLevelTable.USER_EMAIL.Eq(user.Email),
		ExercisePermissionsLevelTable.EXERCISE_IDENTIFIER.Eq(exercise.Id)).
		Into(
		ExercisePermissionsLevelTable.To(&permissionsMap)).
		FetchOne(session)

	if err != nil {
		return nil, err
	}
	return &permissionsMap, nil
}

// FindOrInitSettingsForExercise searches for settings inside cassandra, or
// otherwise inits them
func FindOrInitSettingsForExercise(cassandra *gocql.ClusterConfig, exerciseID gocql.UUID) (map[string]string, error) {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	var eMap map[string]string
	var settings sitrep.SettingsByExerciseIdentifier

	_, err := ctx.Select().
		From(SettingsByExerciseIdentifierTable).
		Where(
		SettingsByExerciseIdentifierTable.ID.Eq(exerciseID)).
		Into(
		SettingsByExerciseIdentifierTable.To(&settings)).
		FetchOne(session)

	if err != nil {
		return eMap, nil
	}
	if len(settings.Settings) > 1 {
		return settings.Settings, nil
	}

	defaultSettings := map[string]string{
		"backgroundColorMenuBar": "#ccc",
		"fontColorMenuBar":       "#555",
		"newsStationEnabled":     "true",
		"twitterEnabled":         "true",
		"facebookEnabled":        "false",
		"youtubeEnabled":         "false",
		"usaidEnabled":           "false",
		"dosEnabled":             "true",
		"contactEnabled":         "true",
		"contactDestination":     "sitrep@vatcinc.com",
		"arcgisMainMapLink":      "",
		"arcgisEmbed":            "true",
	}
	if err := ctx.Upsert(SettingsByExerciseIdentifierTable).
		SetStringStringMap(SettingsByExerciseIdentifierTable.SETTINGS, defaultSettings).
		Where(
		SettingsByExerciseIdentifierTable.ID.Eq(exerciseID)).
		Exec(session); err != nil {
		return eMap, err
	}
	return defaultSettings, nil
}

// UpdateExerciseSetting updates a single setting key and value
func UpdateExerciseSetting(cassandra *gocql.ClusterConfig, exerciseID gocql.UUID, settings map[string]string) (map[string]string, error) {
	session, ctx, _ := WithSession(cassandra)
	defer session.Close()
	ex, err := FindOrInitSettingsForExercise(cassandra, exerciseID)
	if err != nil {
		return ex, err
	}
	if err := mergo.Merge(&settings, ex); err != nil {
		return ex, err
	}
	if err := ctx.Upsert(SettingsByExerciseIdentifierTable).
		SetStringStringMap(SettingsByExerciseIdentifierTable.SETTINGS, settings).
		Where(
		SettingsByExerciseIdentifierTable.ID.Eq(exerciseID)).
		Exec(session); err != nil {
		return ex, err
	}
	return settings, nil

}
