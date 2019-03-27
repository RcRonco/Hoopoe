package dnsproxy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type stringMatchingFunc func(string, string) bool

func matchingFuncMap(action int8) (error, stringMatchingFunc) {
	switch action {
	case PREFIX:
		return nil, func(source string, pattern string) bool {
			return strings.HasPrefix(source, pattern)
		}
	case SUFFIX:
		return nil, func(source string, pattern string) bool {
			return strings.HasSuffix(source, pattern)
		}
	case SUBSTRING:
		return nil, func(source string, pattern string) bool {
			return strings.Contains(source, pattern)
		}
	case REGEXP:
		return nil, func(source string, pattern string) bool {
			if matching, err := regexp.MatchString(pattern, source); err != nil {
				return matching
			} else {
				return false
			}
		}
	default:
		return errors.New("unknown rewrite action"), nil
	}
}

// RULE-TYPE ACTION PATTERN OPTIONS
type MatchingRule struct {
	//options allowRuleOptions
	matchingRule stringMatchingFunc

	Action  int8
	Pattern string
}

func (r *MatchingRule) Parse(rawRule []string) error {
	if len(rawRule) < 3 {
		return fmt.Errorf("%s definition must have 3 fields", rawRule[RuleTypeOffset])
	}

	if val, ok := ActionMap[rawRule[ActionOffset]]; ok {
		r.Action = val
	} else {
		return fmt.Errorf("action %s not supported", rawRule[ActionOffset])
	}

	r.Pattern = rawRule[PatternOffset]

	// TODO: Find a way to compile regexp before server starts
	// Validate rules compiled before start running
	if r.Action == REGEXP {
		if _, err := regexp.Compile(r.Pattern); err == nil {
			return fmt.Errorf("failed to parse Rewrite rule Regexp: %s", err)
		}
	}

	if err, fnc := matchingFuncMap(r.Action); err == nil {
		r.matchingRule = fnc
	} else {
		return fmt.Errorf("rewrite function not found, Action: %s\tMessage: %s",rawRule[ActionOffset], err)
	}


	return nil
}

func (r *MatchingRule) Apply(name string) (bool, string) {
	return r.matchingRule(name, r.Pattern), name
}

func NewMatchingRule(rawRule []string) (error, *MatchingRule) {
	r := new(MatchingRule)

	if err := r.Parse(rawRule); err != nil {
		return err, nil
	} else {
		return nil, r
	}
}

