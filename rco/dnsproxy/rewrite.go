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
	SubstringReplacements	int
}

type rewriteFunc func(*RewriteRule, string) (bool, string)

// RULE-TYPE ACTION PATTERN REPLACEMENT OPTIONS
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
	if len(rawRule) < 4 {
		return errors.New("rewrite definition must have 4 fields")
	}

	if val, ok := ActionMap[rawRule[ActionOffset]]; ok {
		r.Action = val
	} else {
		return fmt.Errorf("action %s not supported", rawRule[ActionOffset])
	}

	r.Pattern = rawRule[PatternOffset]
	r.Replacement = rawRule[ReplacementOffset]
	if !strings.HasSuffix(r.Replacement , ".") {
		r.Replacement  += "."
	}
	if r.Action == REGEXP {
		if pattern, err := regexp.Compile(r.Pattern); err != nil {
			r.Regex = pattern
		} else {
			return fmt.Errorf("failed to parse Rewrite rule Regexp: %s", err)
		}
	}

	if err, fnc := rewriteFuncMap(r.Action); err == nil {
		r.rwRule = fnc
	} else {
		return fmt.Errorf("rewrite function not found, Action: %s\tMessage: %s",rawRule[ActionOffset], err)
	}


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
