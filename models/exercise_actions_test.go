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
