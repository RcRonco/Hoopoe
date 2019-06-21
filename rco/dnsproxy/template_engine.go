package dnsproxy

import (
	"github.com/miekg/dns"
	"net"
	"regexp"
	"strings"
)

const (
	ExprLeftover = `(^\.|^-|\.-|-\.|\.\.|--|-$)`
)

type TemplateEngine struct {
	regionMap *RegionMap
	templateLeftoverRegex *regexp.Regexp
}

func NewTemplateEngine(regionMap *RegionMap) *TemplateEngine {
	tr := new(TemplateEngine)
	tr.regionMap = regionMap
	tr.templateLeftoverRegex = regexp.MustCompile(ExprLeftover)
	return tr
}

func (tr *TemplateEngine) Replace(name string, reqIP net.Addr, req *dns.Msg) (int8, string) {
	if !strings.Contains(name, "{:") {
		return BLOCKED, name
	}

	region := tr.regionMap.GetRegion(strings.Split(reqIP.String(), ":")[0])
	if region == "" {
		name = tr.templateLeftoverRegex.ReplaceAllString(name, "")
	} else {
		// Replace all region templates with the client region
		name = strings.Replace(name, "{REGION}", region, 0)
	}

	return ALLOWED, name
}

