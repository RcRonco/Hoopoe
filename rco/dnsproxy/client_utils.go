package dnsproxy

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
)

type RegionDef struct {
	Region   string   `mapstructure:"region"`
	Networks []string `mapstructure:"networks"`
}

type RegionsDef struct {
	Regions []RegionDef `mapstructure:"regions"`
}

type Region struct {
	Region   string
	Networks []*net.IPNet
}

type RegionMap map[string]Region



func NewRegionMap(path string) (error, RegionMap) {
	regionsDef := new(RegionsDef)
	regionMap := make(RegionMap)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	// Read the client map file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		goto loadSubnetError
	}

	// Parse the YAML file
	if err = yaml.Unmarshal(data, regionsDef); err != nil {
		goto loadSubnetError
	}
	for _, def := range regionsDef.Regions {
		region := Region{Region: def.Region}
		for _, regNet := range def.Networks {
			_, ipnet, err := net.ParseCIDR(regNet)
			if err != nil {
				goto loadSubnetError
			}
			region.Networks = append(region.Networks, ipnet)
		}
		regionMap[region.Region] = region
	}

	if overlapping, regionsOverlapping := regionMap.validateOverlapping(); overlapping {
		return fmt.Errorf("there are overlapping regions: %s", regionsOverlapping), nil
	}

	return nil, regionMap

loadSubnetError:
	return fmt.Errorf("failed to create client map: %s", err), nil
}

func (r *Region)IsRegionsOverlapping(region *Region) bool {
	for _, regNet1 := range r.Networks {
		for _, regNet2 := range region.Networks {
			if regNet1.Contains(regNet2.IP) || regNet2.Contains(regNet1.IP) {
				return true
			}
		}
	}

	return false
}

func (r *Region)IsClientInRegion(ipaddr string) bool {
	for _, regionNet := range r.Networks {
		if regionNet.Contains(net.ParseIP(ipaddr)) {
			return true
		}
	}

	return false
}

func (regionMap *RegionMap)validateOverlapping() (bool, string) {
	for name, region := range *regionMap {
		for name2, region2 := range *regionMap {
			if name != name2 {
				if region.IsRegionsOverlapping(&region2) {
					return true, fmt.Sprintf("%s, %s", name, name2)
				}
			}
		}
	}

	return false, ""
}
func (rm *RegionMap) GetRegion(ipaddr string) string {
	for regionName, region := range *rm {
		if region.IsClientInRegion(ipaddr) {
			return regionName
		}
	}

	return ""
}