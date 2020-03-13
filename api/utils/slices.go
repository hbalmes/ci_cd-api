package utils

import "github.com/herbal828/ci_cd-api/api/models"

//Returns if a slice contains the string.
func Contains(s []models.RequireStatusCheck, e string) bool {
	for _, a := range s {
		if a.Check == e {
			return true
		}
	}
	return false
}
