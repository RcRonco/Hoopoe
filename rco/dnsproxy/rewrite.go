package dnsproxy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	ReplacementOffset = 3
)


type rewriteOptions struct {
	SubstringReplacements int
}

type rewriteFunc func(*RewriteRule, string) (bool, string)

// RULE-TYPE ACTION PATTERN REPLACEMENT OPTIONS...
type RewriteRule struct {
	options rewriteOptions
	rwRule  rewriteFunc

	Action      int8
	Pattern     string
	Replacement string
	Regex       *regexp.Regexp
}

func NewRewriteRule(rawRule []string) (error, *RewriteRule) {
	rw := new(RewriteRule)
	if err := rw.Parse(rawRule); err != nil {
		return err, nil
	} else {
		return nil, rw
	}
}

func (r *RewriteRule) Parse(rawRule []string) error {
	// Validate the number of parameters in the rule definition
	if len(rawRule) < ReplacementOffset + 1 {
		return fmt.Errorf("rewrite definition must have at least %d fields", ReplacementOffset + 1)
	}

	// Validate action type
	if val, ok := ActionMap[rawRule[ActionOffset]]; ok {
		r.Action = val
	} else {
		return fmt.Errorf("action %s not supported", rawRule[ActionOffset])
	}

	// Validate the rule patterns match DNS standard + templating
	if matched, _ := regexp.MatchString(DNS_REGEXPR, rawRule[PatternOffset]); !matched {
		return fmt.Errorf("rewrite pattern must be valid dns string: %s", rawRule[PatternOffset])
	}
	if matched, _ := regexp.MatchString(DNS_REGEXPR, rawRule[ReplacementOffset]); !matched {
		return fmt.Errorf("rewrite replacement must be valid dns string: %s", rawRule[ReplacementOffset])
	}
	// Validate the templating is written correctly brackets
	if !ValidateTemplateBrackets(rawRule[PatternOffset]) {
		return fmt.Errorf("replacement with template must be valid templating: %s", rawRule[PatternOffset])
	}
	if !ValidateTemplateBrackets(rawRule[ReplacementOffset]) {
		return fmt.Errorf("replacement with template must be valid templating: %s", rawRule[ReplacementOffset])
	}

	// Build rule
	r.Pattern = rawRule[PatternOffset]
	r.Replacement = rawRule[ReplacementOffset]
	// Add . suffix
	if !strings.HasSuffix(r.Replacement , ".") {
		r.Replacement  += "."
	}

	// Compile Regex pattern
	if r.Action == REGEXP {
		if pattern, err := regexp.Compile(r.Pattern); err != nil {
			r.Regex = pattern
		} else {
			return fmt.Errorf("failed to parse Rewrite rule Regexp: %s", err)
		}
	}

	// Build Function Map
	if err, fnc := rewriteFuncMap(r.Action); err == nil {
		r.rwRule = fnc
	} else {
		return fmt.Errorf("rewrite function not found, Action: %s\tMessage: %s",rawRule[ActionOffset], err)
	}

	// TODO: Add support for limiting the number of replacement occurs
	r.options.SubstringReplacements = 0

	return nil
}

func (r *RewriteRule) Apply(name string) (bool, string) {
	return r.rwRule(r, name)
}

func rewriteFuncMap(action int8) (error, rewriteFunc) {
	switch action {
	case PREFIX:
		return nil, func(rule *RewriteRule, req string) (bool, string) {
			if strings.HasPrefix(req, rule.Pattern) {
				resp := rule.Replacement + strings.TrimPrefix(req, rule.Pattern)
				return true, resp
			}
			return false, req
		}
	case SUFFIX:
		return nil, func(rule *RewriteRule, req string) (bool, string) {
			if strings.HasSuffix(req, rule.Pattern) {
				resp := strings.TrimSuffix(req, rule.Pattern) + rule.Replacement
				return true, resp
			}
			return false, req
		}
	case SUBSTRING:
		return nil, func(rule *RewriteRule, req string) (bool, string) {
			if strings.Contains(req, rule.Pattern) {
				resp := strings.Replace(req, rule.Pattern, rule.Replacement, rule.options.SubstringReplacements)
				return true, resp
			}
			return false, req
		}
	case REGEXP:
		return nil, func(rule *RewriteRule, req string) (bool, string) {
			if rule.Regex != nil {
				resp := rule.Regex.ReplaceAllString(req, rule.Replacement)
				return true, resp
			}

			return false, req
		}
	default:
		return errors.New("unknown rewrite action"), nil
	}
}
