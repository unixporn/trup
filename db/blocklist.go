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

func AddToBlocklist(moderator, pattern string) error {
	// check if the regex is actually valid
	if _, err := regexp.Compile(pattern); err != nil {
		return err
	}

	if _, err := db.Exec(context.Background(), "INSERT INTO blocked_regexes (pattern, added_by) VALUES ($1, $2)", pattern, moderator); err != nil {
		return err
	}
	if err := addPatternToCache(pattern); err != nil {
		return err
	}
	return nil
}

func RemoveFromBlocklist(pattern string) (removed bool, err error) {
	commandTag, err := db.Exec(context.Background(), "DELETE FROM blocked_regexes WHERE pattern = $1 RETURNING *;", pattern)
	if err != nil {
		return false, err
	}
	if err = updatePatternCache(); err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}

func FindBlockedWordMatch(message string) (string, error) {
	blockRegex, err := getBlockRegex()
	if err != nil {
		return "", err
	}

	submatch := blockRegex.FindString(message)
	return submatch, nil
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
