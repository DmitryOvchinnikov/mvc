package models

import "strings"

const (
	// ErrNotFound is returned when a resource cannot be found in the database.
	ErrNotFound     modelError = "models: resource not found"
	ErrTextRequired modelError = "models: text is required"

	// ErrIDInvalid is returned when an invalid ID is provided to a method like Update.
	ErrIDInvalid privateError = "models: ID provided was invalid"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

type privateError string

func (e privateError) Error() string {
	return string(e)
}
