package dnsproxy

import (
	"errors"
	"regexp"
	"strings"
)

const (
	ExprLeftover = `(^\.|^-|\.-|-\.|\.\.|--|-$)`
)

type TemplateEngine struct {
	templateLeftoverRegex *regexp.Regexp
}

func NewTemplateEngine() *TemplateEngine {
	te := new(TemplateEngine)
	te.templateLeftoverRegex = regexp.MustCompile(ExprLeftover)
	return te
}

func (te *TemplateEngine) Name() string {
	return "TemplatesEngine"
}

func (te *TemplateEngine) Apply(query *EngineQuery, metadata RequestMetadata) (*EngineQuery, error) {
	result := new(EngineQuery)
	result.Queries = query.Queries
	result.dnsMsg = query.dnsMsg
	if len(query.Queries) <= 0 {
		return nil, errors.New("can't get as input an empty EngineQuery")
	}

	name := query.Queries[0].Name
	// Check if template exists in current query
	if !strings.Contains(name, "{") {
		result.Result = ALLOWED
		return result, nil
	}

	if metadata.Region == "" {
		name = te.templateLeftoverRegex.ReplaceAllString(name, "")
	} else {
		// Replace all region templates with the client region
		name = strings.Replace(name, "{REGION}", metadata.Region, 0)
	}

	// Check if the query still contains any template chars ({,})
	if isMatching, err := regexp.MatchString("({|})", name); err != nil || isMatching {
		result.Result = BLOCKED
		return result, err
	}

	result.Queries[0].Name = name
	result.Result = ALLOWED

	return result, nil
}