package dnsproxy

import (
	"fmt"
	"regexp"
	"strings"
	"errors"
)

const (
	PREFIX int8 = iota
	SUFFIX
	SUBSTRING
	REGEXP
)

const (
	RuleTypeOffset = 0
	ActionOffset = 1
	FromOffset = 2
	ToOffset = 3
	OptionOffset = 4
)



var (
	ActionMap = map[string]int8{
		"PREFIX": PREFIX,
		"SUfFIX": SUFFIX,
		"SUBSTRING": SUBSTRING,
		"REGEX": REGEXP,
	}
)

type rewriteOptions struct {
	SubstringReplacements	int
}

type RewriteRule struct {
	Action int8
	From   string
	To     string
	Regex      *regexp.Regexp
	options rewriteOptions
}

type rewriteFunc func(string, RewriteRule) (bool, string)

func (rw *RewriteRule) Parse(rawRule []string) error {
	if len(rawRule) < 4 {
		return errors.New("Rewrite definition must have 4 fields")
	}

	if val, ok := ActionMap[rawRule[ActionOffset]]; ok {
		rw.Action = val
	} else {
		return fmt.Errorf("action %s not supported", rawRule[ActionOffset])
	}

	rw.From = rawRule[FromOffset]
	rw.To = rawRule[ToOffset]
	if rw.Action == REGEXP {
		if pattern, err := regexp.Compile(rw.From); err != nil {
			rw.Regex = pattern
		} else {
			return fmt.Errorf("failed to parse rewrite rule: %s", err)
		}
	}

	return nil
}


func rewriteFuncMap(action int8) (error, rewriteFunc) {
	switch action {
	case PREFIX:
		return nil, func(req string, rule RewriteRule) (bool, string) {
			if strings.HasPrefix(req, rule.From) {
				resp := rule.To + strings.TrimPrefix(req, rule.From)
				return true, resp
			}
			return false, req
		}
	case SUFFIX:
		return nil, func(req string, rule RewriteRule) (bool, string) {
			if strings.HasSuffix(req, rule.From) {
				resp := strings.TrimSuffix(req, rule.From) + rule.To
				return true, resp
			}
			return false, req
		}
	case SUBSTRING:
		return nil, func(req string, rule RewriteRule) (bool, string) {
			if strings.Contains(req, rule.From) {
				resp := strings.Replace(req, rule.From, rule.To, rule.options.SubstringReplacements)
				return true, resp
			}
			return false, req
		}
	case REGEX:
		return nil, func(req string, rule RewriteRule) (bool, string) {
			resp := rule.Regex.ReplaceAllString(req, rule.To)
			return true, resp
		}
	default:
		return errors.New("unknown rewrite action"), nil
	}
}


