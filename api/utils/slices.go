package utils

import "github.com/hbalmes/ci_cd-api/api/models"

//Returns if a slice contains a requireStatusCheck.
func ContainsStatusChecks(s []models.RequireStatusCheck, e string) bool {
	for _, a := range s {
		if a.Check == e {
			return true
		}
	}
	return false
}

//Returns if a slice of string contains an string
func StringContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
