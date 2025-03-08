package util

import "regexp"

func IsValidSHA512(hash string) bool {
	if len(hash) != 128 {
		return false
	}

	match, _ := regexp.MatchString("^[a-fA-F0-9]{128}$", hash)
	return match
}
