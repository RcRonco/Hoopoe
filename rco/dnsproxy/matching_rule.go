package dnsproxy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type stringMatchingFunc func(*MatchingRule, string) bool

func matchingFuncMap(action int8) (error, stringMatchingFunc) {
	switch action {
	case PREFIX:
		return nil, func(rule *MatchingRule, source string) bool {
			return strings.HasPrefix(source, rule.Pattern)
		}
	case SUFFIX:
		return nil, func(rule *MatchingRule, source string) bool {
			return strings.HasSuffix(source, rule.Pattern)
		}
	case SUBSTRING:
		return nil, func(rule *MatchingRule, source string) bool {
			return strings.Contains(source, rule.Pattern)
		}
	case REGEXP:
		return nil, func(rule *MatchingRule, source string) bool {
			return rule.Regex.MatchString(source)
		}
	default:
		return errors.New("unknown matching action"), nil
	}
}

// RULE-TYPE ACTION PATTERN OPTIONS
type MatchingRule struct {
	//options allowRuleOptions
	matchingRule stringMatchingFunc

	Action  int8
	Pattern string
	Regex *regexp.Regexp
}

func (r *MatchingRule) Parse(rawRule []string) error {
	// Validate the number of parameters in the rule definition
	if len(rawRule) < PatternOffset + 1 {
		return fmt.Errorf("%s definition must have at least %d fields",
						  rawRule[RuleTypeOffset], PatternOffset + 1)
	}

	// Validate action type
	if val, ok := ActionMap[rawRule[ActionOffset]]; ok {
		r.Action = val
	} else {
		return fmt.Errorf("action %s not supported", rawRule[ActionOffset])
	}

	// Validate the rule patterns match DNS standard + templating
	if matched, _ := regexp.MatchString(DNS_REGEXPR, rawRule[PatternOffset]); !matched {
		return fmt.Errorf("pattern must be valid dns string: %s", rawRule[PatternOffset])
	}

	// Validate the templating is written correctly brackets
	if !ValidateTemplateBrackets(rawRule[PatternOffset]) {
		return fmt.Errorf("pattern with template must be valid templating: %s", rawRule[PatternOffset])
	}

	r.Pattern = rawRule[PatternOffset]

	// Validate rules compiled before start running
	if r.Action == REGEXP {
		if regex, err := regexp.Compile(r.Pattern); err != nil {
			return fmt.Errorf("failed to parse Rewrite rule Regexp: %s", err)
		} else {
			r.Regex = regex
		}
	}

	// Build Function Map
	if err, fnc := matchingFuncMap(r.Action); err == nil {
		r.matchingRule = fnc
	} else {
		return fmt.Errorf("rewrite function not found, Action: %s\tMessage: %s",rawRule[ActionOffset], err)
	}

	return nil
}

func (r *MatchingRule) Apply(name string) (bool, string) {
	return r.matchingRule(r, name), name
}

func NewMatchingRule(rawRule []string) (error, *MatchingRule) {
	r := new(MatchingRule)

	if err := r.Parse(rawRule); err != nil {
		return err, nil
	} else {
		return nil, r
	}
}

