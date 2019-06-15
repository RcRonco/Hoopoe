package dnsproxy

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"net"
	"os"
)

type Region struct {
	Networks []net.IPNet `yaml:"networks"`
	Region  string `yaml:"region"`
}

type RegionMap map[string]Region

func LoadRegionMap(path string) (error, *RegionMap) {
	regionMap := make(RegionMap)
	var err error
	data := make([]byte, 4)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	// Read the client map file
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		goto loadSubnetError
	}
	if _, err = file.Read(data); err != nil {
		goto loadSubnetError
	}

	// Parse the YAML file
	if err = yaml.Unmarshal(data, regionMap); err != nil {
		goto loadSubnetError
	}

	return nil, &regionMap

loadSubnetError:
	return fmt.Errorf("failed to create client map: %s", err), nil
}

func (r *Region)IsClientInRegion(ipaddr string) bool {
	for _, regionNet := range r.Networks {
		if regionNet.Contains(net.ParseIP(ipaddr)) {
			return true
		}
	}

	return false
}

func (rm *RegionMap) GetRegion(ipaddr string) string {
	for regionName, region := range *rm {
		if region.IsClientInRegion(ipaddr) {
			return regionName
		}
	}

	return ""
}