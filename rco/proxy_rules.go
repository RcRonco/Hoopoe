package rco

import (
	"regexp"
	"strings"
)

const (
	Allow = iota
	Deny = iota
	Rewrite = iota
)

type ProxyRule struct {
	Rule int
	Regex *regexp.Regexp
	NewPattern string

}

func CompileRules(conf_rules []ProxyRuleConfig) []ProxyRule {
	rules := []ProxyRule{}
	for _, cr := range conf_rules {
		rule := ProxyRule{}
		switch (cr.Rule) {
			case "Allow", "allow", "A":
				rule.Rule = Allow
				break
			case "Deny", "deny", "D":
				rule.Rule = Deny
				break
			case "Rewrite", "rewrite", "RW":
				rule.Rule = Rewrite
				break
			default:
				return []ProxyRule{}
		}

		if rule.Rule == Rewrite {
			rule.NewPattern = cr.NewPattern
		}
		rule.Regex, _ = regexp.Compile(cr.Pattern)
		rules = append(rules, rule)
	}

	return rules
}

func (pr *ProxyRule) CheckPattern(name string) bool {
	indexes := pr.Regex.FindStringIndex(name)
	return 0 < len(indexes)
}

func (pr *ProxyRule) Apply(name string) string {
	matches := pr.Regex.FindAllString(name, 5)
	new_name := name
	for _, m := range matches {
		new_name  = strings.Replace(new_name, m, pr.NewPattern, 1)
	}

	return new_name
}