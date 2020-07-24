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
	_, err := db.Query(context.Background(), "INSERT INTO blocked_regexes VALUES ($1)", pattern)
	_, _ = getRegexCache()
	regexCache.addPattern(pattern)
	return err
}

func RemoveFromBlocklist(pattern string) (bool, error) {
	rows, err := db.Query(context.Background(), "DELETE FROM blocked_regexes WHERE pattern = $1", pattern)
	if err != nil {
		return false, err
	}
	rows.Next()
	values, err := rows.Values()
	if err != nil {
		return false, err
	}
	return values[0].(int) > 0, nil
}

type RegexCache struct {
	patterns        []string
	combinedPattern regexp.Regexp
}

var regexCache *RegexCache

func (cache *RegexCache) addPattern(pattern string) error {
	cache.patterns = append(cache.patterns, pattern)
	regex, err := regexp.Compile(strings.Join(cache.patterns, "|"))
	if err != nil {
		return err
	}
	cache.combinedPattern = *regex
	return nil
}

func getRegexCache() (*RegexCache, error) {
	if regexCache == nil {
		patterns, err := GetBlocklist()
		if err != nil {
			return nil, err
		}
		regex, err := regexp.Compile(strings.Join(patterns, "|"))
		if err != nil {
			return nil, err
		}

		regexCache = &RegexCache{
			patterns:        patterns,
			combinedPattern: *regex,
		}

	}
	return regexCache, nil
}

func ContainsBlockedWords(message string) (bool, error) {
	cache, err := getRegexCache()
	if err != nil {
		return false, err
	}
	return cache.combinedPattern.MatchString(message), nil
}
