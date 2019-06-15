package dnsproxy

import (
  "errors"
  "fmt"
	"github.com/miekg/dns"
	"net"
	"regexp"
	"strings"
)

const (
	EXPR_LEFOVER = `(^\.|^-|\.-|-\.|\.\.|--|-$)`
)

type TemplateEngine struct {
	regionMap *RegionMap
	templateLeftoverRegex *regexp.Regexp
}

func NewTemplateRule() (error, *TemplateEngine) {
	tr := new(TemplateEngine)
	tr.templateLeftoverRegex = regexp.MustCompile(EXPR_LEFOVER)
	return nil, tr
}

func (tr *TemplateEngine) Replace(name string, reqIP net.Addr, req *dns.Msg) (bool, string) {
	if !strings.Contains(name, "{") {
		return false, name
	}

	region := tr.regionMap.GetRegion(strings.Split(reqIP.String(), ":")[0])
	if region == "" {
		name = tr.templateLeftoverRegex.ReplaceAllString(name, "")
	} else {
		// Replace all region templates with the client region
		name = strings.Replace(name, "{REGION}", region, 0)
	}

	return true, name
}

