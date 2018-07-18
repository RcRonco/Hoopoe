package dnsproxy

import (
	"regexp"
	"strings"
	"github.com/golang/glog"
)

const (
	Allow = iota
	Deny = iota
	Rewrite = iota
)

const (
	NOERR = iota
	BLOCKED = iota
	ERROR = iota
)

type Result struct {
	Code  uint32
	Err string
}

type Rule struct {
	Type       int
	Regex      *regexp.Regexp
	NewPattern string

}

type RuleEngine struct {
	white_list    []Rule
	black_list    []Rule
	rewrite_rules []Rule
}

// Compile rules from configuration
func (re *RuleEngine) CompileRules(conf_rules []ProxyRuleConfig) error {
	var err error
	// Build every rule on the config
	for index, cr := range conf_rules {
		rule := Rule{}

		// Compile regular expression
		rule.Regex,err = regexp.Compile(cr.Pattern)
		if err != nil {
			glog.Exitf("Failed to compile rule regular experssion, rule number: %d\n", index)
			panic(err)
		}

		// Check what type of rule it is and add it to the right array
		switch (strings.ToLower(cr.Rule)) {
		case "allow", "a":
			rule.Type = Allow
			re.white_list = append(re.white_list, rule)
			break
		case "deny", "d":
			rule.Type = Deny
			re.black_list = append(re.black_list, rule)
			break
		case "rewrite", "rw":
			rule.Type = Rewrite
			rule.NewPattern = cr.NewPattern
			re.rewrite_rules = append(re.rewrite_rules, rule)
			break
		default:
			glog.Exitf("Unsupported rule type - \"%s\"in rule number %d", cr.Rule, index)
		}
	}

	return nil
}

// Apply all rules on query
func (re *RuleEngine) Apply(query string) (Result, string) {
	// Copy query to not change original one
	nquery := query

	// Apply all Allow rules
	for _, rule := range re.white_list {
		res,_ := rule.Apply(query)

		// Check if the rule Blocked
		if res.Code != NOERR {
			return res, ""
		}
	}

	// Apply all Deny rules
	for _, rule := range re.black_list {
		res,_ := rule.Apply(query)

		// Check if the rule Blocked
		if res.Code != NOERR {
			return res, ""
		}
	}

	// Apply all rewrite rules
	for _, rule := range re.rewrite_rules {
		_, nquery = rule.Apply(nquery)
	}

	return Result{Code: NOERR, Err: ""}, nquery
}

// Check if the pattern matches a string
func (pr *Rule) CheckPattern(name string) bool {
	indexes := pr.Regex.FindStringIndex(name)
	return 0 < len(indexes)
}

// Wrapper for all applying rule internal functions
func (pr *Rule) Apply(name string) (Result, string) {
	switch pr.Type {
	case Allow:
		return pr.apply_allow(name)
	case Deny:
		return pr.apply_deny(name)
	case Rewrite:
		return pr.apply_rewrite(name)
	default:
		return Result{Code:ERROR, Err: ""}, ""
	}
}

// Applying a Allow rule on string
func (pr *Rule) apply_allow(name string) (Result, string) {
	// Build blocked result
	res := Result{Code: BLOCKED, Err: ""}

	// Check if the name matching the pattern if yes change result to noerr
	if (pr.Regex.MatchString(name)) {
		res.Code = NOERR
	}

	return res,name
}

// Applying a Deny rule on string
func (pr *Rule) apply_deny(name string) (Result, string) {
	// Build noerr result
	res := Result{Code: NOERR, Err: ""}

	// Check if the name matching the pattern if yes change result to blocked
	if (pr.Regex.MatchString(name)) {
		res.Code = BLOCKED
	}

	return res,name
}

// Applying a Rewrite rule on string
func (pr *Rule) apply_rewrite(name string) (Result, string) {
	// Rewrite string and return result
	new_name := pr.Regex.ReplaceAllString(name, pr.NewPattern)
	return Result{Code: NOERR, Err: ""},new_name
}