package models

import "fmt"

// AlreadyExistsError holds a specific error
type AlreadyExistsError struct {
	Message string
}

func (u *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", u.Message)
}

// ValidationError holds a specific error
type ValidationError struct {
	Field  string
	Reason string
}

func (u *ValidationError) Error() string {
	return fmt.Sprintf("%v validation failed: %v", u.Field, u.Reason)
}
