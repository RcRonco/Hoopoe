package dnsproxy

import (
	log "github.com/Sirupsen/logrus"
	"strings"
)

const (
	ALLOWED int8 = 1 << iota
	BLOCKED int8 = 1 << iota
	ERROR   int8 = 1 << iota
)

const (
	RuleTypeOffset = 0
	ActionOffset = 1
	PatternOffset = 2
)

const (
	PassType int8 =  iota
	RewriteType
	AllowType
	DenyType
)

const (
	PREFIX int8 = iota
	SUFFIX
	SUBSTRING
	REGEXP
)

var (
	ActionMap = map[string]int8{
		"PREFIX": PREFIX,
		"SUfFIX": SUFFIX,
		"SUBSTRING": SUBSTRING,
		"REGEX": REGEXP,
	}

	RuleTypeMap = map[string]int8{
		"pass": PassType,
		"p": PassType,
		"rewrite": RewriteType,
		"rw": RewriteType,
		"allow": AllowType,
		"a": AllowType,
		"deny": DenyType,
		"d": DenyType,
	}
)

type Rule interface {
	Parse([]string) error
	Apply(string) (bool, string)
}

type RuleEngine struct {
	rules   map[int8][]Rule
	scanAll bool
}

func (re *RuleEngine) SetScanAll(scanAll bool) {
	re.scanAll = scanAll
}

/*
	Build new engine
	Rule definition format:
	RULETYPE ACTION FROM TO OPTIONS
 */
func NewEngine(rawRules []string) *RuleEngine {
	engine := new(RuleEngine)
	engine.rules = make(map[int8][]Rule)

	for index, rr := range rawRules {
		fields := strings.Fields(strings.ToLower(rr))
		switch fields[RuleTypeOffset] {
			case "rewrite", "rw":
				if err, rw := NewRewriteRule(fields); err != nil {
					log.Fatalf("%d - Failed to parse rewrite rule: %s", index, err)
				} else {
					engine.rules[RewriteType] = append(engine.rules[RewriteType], rw)
				}
				break
			case "pass", "p", "allow", "a", "deny", "d":
				if err, r := NewMatchingRule(fields); err == nil {
					log.Fatalf("%d - Failed to parse rule: %s", index, err)
				} else {
					engine.rules[RuleTypeMap[fields[0]]] = append(engine.rules[RuleTypeMap[fields[0]]], r)
				}
				break
		default:
			log.Fatalf("Unsupported rule type - \"%s\"in rule number %d", fields[0], index)
		}
	}

	return engine
}

func (re *RuleEngine) Apply(query string) (int8, string) {
	// Apply Pass Rules
	for _, mr := range re.rules[PassType] {
		if pass, _ := mr.Apply(query); pass {
			return ALLOWED, query
		}
	}

	// Apply Allow Rules
	var res = BLOCKED
	for _, ar := range re.rules[AllowType] {
		if allowed, _ := ar.Apply(query); allowed {
			res = ALLOWED
			break
		}
	}

	if res == BLOCKED {
		return BLOCKED, ""
	}

	// Apply Deny Rules
	for _, dr := range re.rules[DenyType] {
		if denied, _ := dr.Apply(query); denied {
			return BLOCKED, ""
		}
	}

	// Apply rewrites Rules
	var newQuery = query
	for _, rw := range re.rules[RewriteType] {
		rewrite, result := rw.Apply(query)
		newQuery = result

		// Exit rewrites if scanAll not sets and rewrite applied
		if rewrite && !re.scanAll {
			break
		}
	}

	// If passed allow rules and not blocked by deny or change return ALLOWED with original string
	return ALLOWED, newQuery
}
