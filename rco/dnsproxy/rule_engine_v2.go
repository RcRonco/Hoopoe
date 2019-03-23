package dnsproxy

import (
	"errors"
	"fmt"
	"strings"
)

type IRule interface {
	Parse([]string) error
	Apply() (error, string)
}

const (
	PassType int8 =  iota
	RewriteType
	AllowType
	DenyType
	UnknownType
)



type RuleEngineV2 struct {
	rules   map[int8][]IRule
	scanAll bool
}

/*
	Build new engine
	Rule definition format:
	RULETYPE ACTION FROM TO OPTIONS
 */
func NewEngine(rawRules []string) (error, *RuleEngineV2) {

	for _, rr := range rawRules {
		fields := strings.Fields(strings.ToLower(rr))
		switch fields[RuleTypeOffset] {
			case "rewrite", "rw":


		}
	}
}
