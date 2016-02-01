package models

import (
	"github.com/fkasper/sitrep-authentication/schema"
	"github.com/gocql/gocql"
)

// ExerciseByIdentifierTable is a reference to the users cassandra table
var ExerciseByIdentifierTable = sitrep.ExerciseByIdentifierTableDef()

// UsersInExerciseTable is a reference to the users cassandra table
var UsersInExerciseTable = sitrep.CreateUsersInExerciseTableDef()

// FindExerciseByID receives an exercise from the database
func FindExerciseByID(cassandra *gocql.ClusterConfig, id string) (*sitrep.ExerciseByIdentifier, error) {
	return nil, nil
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
func FindExercisePermissionsForUser(cassandra *gocql.ClusterConfig, user *sitrep.UsersByEmail, exercise *sitrep.ExerciseByIdentifier) (*sitrep.ExercisePermissionsLevel, error) {
	return nil, nil
}
