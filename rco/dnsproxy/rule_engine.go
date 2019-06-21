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
		"SUFFIX": SUFFIX,
		"SUBSTRING": SUBSTRING,
		"REGEXP": REGEXP,
	}

	RuleTypeMap = map[string]int8{
		"PASS": PassType,
		"P": PassType,
		"REWRITE": RewriteType,
		"RW": RewriteType,
		"ALLOW": AllowType,
		"A": AllowType,
		"DENY": DenyType,
		"D": DenyType,
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
func NewRuleEngine(rawRules []string) *RuleEngine {
	engine := new(RuleEngine)
	engine.rules = make(map[int8][]Rule)

	log.Info("Start compiling rulesEngine")

	// Compile every rule definition
	for index, rr := range rawRules {
		// Split all rulesEngine into fields and convert to UPPER case
		fields := strings.Fields(strings.ToUpper(rr))
		if !strings.HasSuffix(fields[PatternOffset], ".") {
			fields[PatternOffset] += "."
		}
		// Parse rulesEngine by type and add to the rule map
		switch fields[RuleTypeOffset] {
			case "REWRITE", "RW":
				if err, rw := NewRewriteRule(fields); err != nil {
					log.Fatalf("%d - Failed to parse rewrite rule: %s", index, err)
				} else {
					engine.rules[RewriteType] = append(engine.rules[RewriteType], rw)
				}
				break
			case "PASS", "P", "ALLOW", "A", "DENY", "D":
				if err, r := NewMatchingRule(fields); err != nil {
					log.Fatalf("%d - Failed to parse rule: %s", index, err)
				} else {
					engine.rules[RuleTypeMap[fields[0]]] = append(engine.rules[RuleTypeMap[fields[0]]], r)
				}
				break
		default:
			log.Fatalf("Unsupported rule type - \"%s\"in rule number %d", fields[0], index)
		}
	}

	log.Info("Compiling rulesEngine ended successfully")
	return engine
}

func (re *RuleEngine) Apply(query string) (int8, string) {
	// Convert query into UPPER case to match all UPPER case rulesEngine
	var newQuery = strings.ToUpper(query)
	// Apply Pass Rules
	for _, mr := range re.rules[PassType] {
		if pass, _ := mr.Apply(newQuery); pass {
			return ALLOWED, query
		}
	}

	// Apply Allow Rules
	var res = BLOCKED
	for _, ar := range re.rules[AllowType] {
		if allowed, _ := ar.Apply(newQuery); allowed {
			res = ALLOWED
			break
		}
	}

	if res == BLOCKED {
		return BLOCKED, ""
	}

	// Apply Deny Rules
	for _, dr := range re.rules[DenyType] {
		if denied, _ := dr.Apply(newQuery); denied {
			return BLOCKED, ""
		}
	}

	// Apply rewrites Rules
	for _, rw := range re.rules[RewriteType] {
		rewrite, result := rw.Apply(newQuery)
		newQuery = result

		// Exit rewrites if scanAll not sets and rewrite applied
		if rewrite && !re.scanAll {
			break
		}
	}

	// If passed allow rulesEngine and not blocked by deny or change return ALLOWED with original string
	return ALLOWED, newQuery
}
