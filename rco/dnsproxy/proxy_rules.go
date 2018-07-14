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

type ProxyRule struct {
	Rule int
	Regex *regexp.Regexp
	NewPattern string

}

func CompileRules(conf_rules []ProxyRuleConfig) []ProxyRule {
	rules := []ProxyRule{}
	var err error
	for index, cr := range conf_rules {
		rule := ProxyRule{}
		switch (strings.ToLower(cr.Rule)) {
			case "allow", "a":
				rule.Rule = Allow
				break
			case "deny", "d":
				rule.Rule = Deny
				break
			case "rewrite", "rw":
				rule.Rule = Rewrite
				break
			default:
				glog.Exitf("Unsupported rule type - \"%s\"in rule number %d", cr.Rule, index)
		}

		if rule.Rule == Rewrite {
			rule.NewPattern = cr.NewPattern
		}

		rule.Regex,err = regexp.Compile(cr.Pattern)
		if err != nil {
			glog.Errorf("Failed to compile rule number: %d\n", index)
			panic(err)
		}
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