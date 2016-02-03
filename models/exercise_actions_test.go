package models_test

import (
	"testing"
	"time"

	"github.com/fkasper/sitrep-authentication/models"
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
)

// ExerciseByIdentifierTable is a reference to the users cassandra table
var ExerciseByIdentifierTable = sitrep.ExerciseByIdentifierTableDef()

// UsersInExerciseTable is a reference to the users cassandra table
var UsersInExerciseTable = sitrep.CreateUsersInExerciseTableDef()

// ExercisePermissionsLevelTable is a reference to the users cassandra table
var ExercisePermissionsLevelTable = sitrep.ExercisePermissionsLevelTableDef()

// ActiveUntil time.Time
//
// ExerciseDescription string
//
// ExerciseName string
//
// HasActivation bool
//
// Id gocql.UUID
//
// IsActive bool

func mockExercise() *sitrep.ExerciseByIdentifier {
	uuid, err := gocql.ParseUUID("582412e9-6e2b-493f-b5ed-889f42584861")
	if err != nil {
		panic(err)
	}
	return &sitrep.ExerciseByIdentifier{
		Id:                  uuid,
		ExerciseDescription: "Beta Exercise Description",
		ExerciseName:        "Beta Exercise",
		HasActivation:       true,
		IsActive:            true,
		ActiveUntil:         time.Now().Add(time.Hour * 72),
	}
}
func initExercise(exercise *sitrep.ExerciseByIdentifier) {
	var ex *sitrep.ExerciseByIdentifier
	if exercise != nil {
		ex = exercise
	} else {
		ex = mockExercise()
	}
	session, ctx := mockDb()
	defer session.Close()
	err := ctx.Store(ExerciseByIdentifierTable.Bind(*ex)).Exec(session)
	if err != nil {
		panic(err)
	}
}

func addUserToExercise(user *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier) {
	mapping := map[string]string{}
	mapping[exercise.Id.String()] = exercise.ExerciseName
	membership := &sitrep.CreateUsersInExercise{
		Email:     user.Email,
		Exercises: mapping,
	}
	session, ctx := mockDb()
	defer session.Close()
	err := ctx.Store(UsersInExerciseTable.Bind(*membership)).Exec(session)
	if err != nil {
		panic(err)
	}
}

func addUserPermissionToExercise(user *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier, admin bool, oc bool, trainee bool) {
	membership := &sitrep.ExercisePermissionsLevel{
		UserEmail:          user.Email,
		ExerciseIdentifier: exercise.Id,
		IsOc:               oc,
		IsAdmin:            admin,
		IsTrainee:          trainee,
		RoleDescription:    "Demo Role",
	}
	session, ctx := mockDb()
	defer session.Close()
	err := ctx.Store(ExercisePermissionsLevelTable.Bind(*membership)).Exec(session)
	if err != nil {
		panic(err)
	}
}

// TESTS

func TestExercise_Find(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)

	_, err := models.FindExerciseByID(dbConn(), exercise.Id)
	if err != nil {
		t.Fatalf("Exercise was not found in the Database oO?")
	}
}

func TestUser_ForExercise(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)
	addUserToExercise(user, exercise)
	exMap, err := models.FindExercisesForUser(dbConn(), user)
	if err != nil {
		t.Fatalf("No exercises were found for the user")
	}

	if exMap.Exercises[exercise.Id.String()] != exercise.ExerciseName {
		t.Fatalf("Invalid Email found")
	}

}

func Test_Exercise_Permission_Selection(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)
	addUserToExercise(user, exercise)
	addUserPermissionToExercise(user, exercise, true, true, true)

	permissions, err := models.FindExercisePermissionsForUser(dbConn(), user, exercise)
	if err != nil {
		t.Fatalf("No permissions were found for user %s", user.RealName)
	}

	if !permissions.IsAdmin {
		t.Fatalf("No Admin permissions found, while they were set")
	}

	if !permissions.IsOc {
		t.Fatalf("No OC permissions found, while they were set")
	}

	if !permissions.IsTrainee {
		t.Fatalf("No Trainee permissions found, while they were set")
	}

}

func Test_Exercise_Empty_User(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)
	addUserToExercise(user, exercise)
	addUserPermissionToExercise(user, exercise, true, true, true)

	_, err := models.FindExercisePermissionsForUser(dbConn(), nil, exercise)
	if err == nil {
		t.Fatalf("Test succeeded, but should fail. No User was supplied")
	}

}

func Test_Exercise_Empty_Exercise(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)
	addUserToExercise(user, exercise)
	addUserPermissionToExercise(user, exercise, true, true, true)

	_, err := models.FindExercisePermissionsForUser(dbConn(), user, nil)
	if err == nil {
		t.Fatalf("Test succeeded, but should fail. No Exercise was supplied")
	}

}
func Test_Exercise_Settings_Init(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)

	_, err := models.FindOrInitSettingsForExercise(dbConn(), exercise.Id)
	if err != nil {
		t.Fatalf("Settings fetch failed")
	}

}

func Test_Exercise_Settings_Successful(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)

	settings, err := models.FindOrInitSettingsForExercise(dbConn(), exercise.Id)
	if err != nil {
		t.Fatalf("Settings fetch failed")
	}

	if settings["backgroundColorMenuBar"] != "#ccc" {
		t.Fatalf("Settings check failed! %v", settings)
	}

}

//
func Test_Exercise_Settings_Update(t *testing.T) {
	user := mockUser()
	initUser(user)
	exercise := mockExercise()
	initExercise(exercise)
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
		"key":                    "value",
	}
	settings, err := models.UpdateExerciseSetting(dbConn(), exercise.Id, defaultSettings)
	if err != nil {
		t.Fatalf("Settings update failed")
	}

	if settings["key"] != "value" {
		t.Fatalf("Settings check failed! %v", settings)
	}

	settings2, err := models.FindOrInitSettingsForExercise(dbConn(), exercise.Id)
	if err != nil {
		t.Fatalf("Settings fetch failed")
	}
	if settings2["key"] != "value" {
		t.Fatalf("Settings check failed! %v", settings2)
	}

}
