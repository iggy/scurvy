package errors

import (
	"log"
)

// CheckErr - check error code and pani
// This is mostly to deduplicate code across all the files
func CheckErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
