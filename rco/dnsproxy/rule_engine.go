package dnsproxy

import (
	"regexp"
	"strings"
	log "github.com/Sirupsen/logrus"
)

const (
	AllowRule   int8 = 1 << iota
	DenyRule    int8 = 1 << iota
	RewriteRule int8 = 1 << iota
	PassRule    int8 = 1 << iota
)

const (
	NOERR   int8 = 1 << iota
	SKIP    int8 = 1 << iota
	ALLOWED int8 = 1 << iota
	BLOCKED int8 = 1 << iota
	ERROR   int8 = 1 << iota
)

type Result struct {
	Code  int8
	Err string
}

type Rule struct {
	Type       int8
	Regex      *regexp.Regexp
	NewPattern string
	Expression string
}

type RuleEngine struct {
	passList     []Rule
	whiteList    []Rule
	blackList    []Rule
	rewriteRules []Rule
	scanAll      bool
}

func (re *RuleEngine) SetScanAll(enable bool) {
	re.scanAll = enable
}

// Compile rules from configuration
func (re *RuleEngine) CompileRules(confRules []ProxyRuleConfig) error {
	var err error

	// Build every rule on the config
	for index, cr := range confRules {
		rule := Rule{}

		// Compile regular expression
		rule.Regex,err = regexp.Compile(cr.Pattern)
		rule.Expression = cr.Pattern
		if err != nil {
			log.Fatalf("Failed to compile rule regular experssion, rule number: %d\n", index)
			panic(err)
		}

		// Check what type of rule it is and add it to the right array
		switch strings.ToLower(cr.Rule) {
		case "pass", "p":
			rule.Type = PassRule
			re.passList = append(re.passList, rule)
			break
		case "allow", "a":
			rule.Type = AllowRule
			re.whiteList = append(re.whiteList, rule)
			break
		case "deny", "d":
			rule.Type = DenyRule
			re.blackList = append(re.blackList, rule)
			break
		case "rewrite", "rw":
			rule.Type = RewriteRule
			rule.NewPattern = cr.NewPattern
			re.rewriteRules = append(re.rewriteRules, rule)
			break
		default:
			log.Fatalf("Unsupported rule type - \"%s\"in rule number %d", cr.Rule, index)
		}
	}

	return nil
}

// Apply all rules on query
func (re *RuleEngine) Apply(query string) (Result, string) {
	// Apply passthrough rules
	for _, rule := range re.passList {
		res, _ := rule.Apply(query)
		if res.Code == NOERR {
			return Result{Code: NOERR, Err: ""}, query
		}
	}

	var result = Result{BLOCKED, ""}

	// Apply AllowRule rules
	for _, rule := range re.whiteList {
		result, _ = rule.Apply(query)

		// Check if the rule Blocked
		if result.Code == ALLOWED { break }
	}

	// Check if the query blocked by whitelist rule
	if len(re.whiteList) > 0 && result.Code == BLOCKED { return result, "" }
	if result.Code == ERROR { return result, result.Err }

	// Apply DenyRule rules
	for _, rule := range re.blackList {
		res, _ := rule.Apply(query)

		// Check if the rule Blocked
		if res.Code == BLOCKED {
			return res, ""
		}
	}

	nquery := query
	// Apply all rewrite rules
	for _, rule := range re.rewriteRules {
		result, nquery = rule.Apply(nquery)
		if result.Code == NOERR && re.scanAll == false {
			break
		}
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
	case PassRule:
		return pr.applyPass(name)
	case AllowRule:
		return pr.applyAllow(name)
	case DenyRule:
		return pr.applyDeny(name)
	case RewriteRule:
		return pr.applyRewrite(name)
	default:
		return Result{Code:ERROR, Err: ""}, ""
	}
}

// Applying a AllowRule rule on string
func (pr *Rule) applyPass(name string) (Result, string) {
	// Build blocked result
	res := Result{Code: NOERR, Err: ""}

	// Check if the name matching the pattern if yes change result to noerr
	if pr.Regex.MatchString(name) {
		res.Code = SKIP
	}

	return res, name
}

// Applying a AllowRule rule on string
func (pr *Rule) applyAllow(name string) (Result, string) {
	// Build blocked result
	res := Result{Code: BLOCKED, Err: ""}

	// Check if the name matching the pattern if yes change result to noerr
	if pr.Regex.MatchString(name) {
		res.Code = ALLOWED
	}

	return res, name
}

// Applying a DenyRule rule on string
func (pr *Rule) applyDeny(name string) (Result, string) {
	// Build noerr result
	res := Result{Code: ALLOWED, Err: ""}

	// Check if the name matching the pattern if yes change result to blocked
	if pr.Regex.MatchString(name) {
		res.Code = BLOCKED
	}

	return res, name
}

// Applying a RewriteRule rule on string
func (pr *Rule) applyRewrite(name string) (Result, string) {
	// RewriteRule string and return result
	newName := pr.Regex.ReplaceAllString(name, pr.NewPattern)
	return Result{Code: SKIP, Err: ""}, newName
}