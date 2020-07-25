package db

import (
	"context"
	"regexp"
	"strings"
)

func GetBlocklist() ([]string, error) {
	rows, err := db.Query(context.Background(), "SELECT pattern from blocked_regexes")
	if err != nil {
		return nil, err
	}
	patterns := []string{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, values[0].(string))
	}
	return patterns, nil
}

func AddToBlocklist(pattern string) error {
	// check if the regex is actually valid
	_, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	_, err = db.Query(context.Background(), "INSERT INTO blocked_regexes VALUES ($1)", pattern)
	if err != nil {
		return err
	}
	addPatternToCache(pattern)
	return nil
}

func RemoveFromBlocklist(pattern string) (bool, error) {
	rows, err := db.Query(context.Background(), "DELETE FROM blocked_regexes WHERE pattern = $1 RETURNING *;", pattern)
	if err != nil {
		return false, err
	}
	// apparently, the database will only be updated after this is ran
	hasNext := rows.Next()
	updatePatternCache()

	return hasNext, nil
}

func ContainsBlockedWords(message string) (bool, error) {
	blockRegex, err := getBlockRegex()
	return blockRegex.MatchString(message), err
}

var patternCache *regexp.Regexp

func updatePatternCache() error {
	patterns, err := GetBlocklist()
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		// this regex should never match anything.
		// If the blocklist is empty, we need to make sure nothing is deleted.
		patternCache = regexp.MustCompile("a^")
	} else {
		regex, err := compileCombinedPatternRegex(patterns)
		if err != nil {
			return err
		}
		patternCache = regex
	}
	return nil
}

func addPatternToCache(pattern string) error {
	patterns, err := GetBlocklist()
	if err != nil {
		return err
	}
	patterns = append(patterns, pattern)
	regex, err := compileCombinedPatternRegex(patterns)
	if err != nil {
		return err
	}
	patternCache = regex
	return nil
}

func getBlockRegex() (*regexp.Regexp, error) {
	if patternCache == nil {
		err := updatePatternCache()
		if err != nil {
			return nil, err
		}
	}
	return patternCache, nil
}

func compileCombinedPatternRegex(patterns []string) (*regexp.Regexp, error) {
	regex, err := regexp.Compile("(?i)" + strings.Join(patterns, "|"))
	return regex, err
}
