package dnsproxy

import (
	"github.com/Sirupsen/logrus"
	"net"
	"testing"
)

func TestLoadClientMap(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(code int){}
	err, cm := NewRegionMap("test_data/client_map_not_exists_aaaa.yaml")
	if err == nil {
		t.Error("RegionMap won't fail when trying to access non-existing file")
	}
	if cm != nil {
		t.Error("RegionMap return not nil while trying to access non-existing file")
	}
}

// ---
//regions:
//  - region: il
//    networks:
//    - "192.168.1.0/24"
//  - region: us
//    networks:
//    - "10.0.3.0/14"
//    - "1.1.1.1/32"

func TestLoadClientMapRegular(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(code int){}
	err, rm := NewRegionMap("test_data/clientMap.yaml")
	if err != nil {
		t.Error("RegionMap failed when trying to access client file")
	}
	if rm == nil {
		t.Error("RegionMap return nil while trying to access client file")
	}
	if region, ok := rm["il"]; ok {
		_, ipnet, _ := net.ParseCIDR("192.168.1.0/24")
		if region.Region != "il" || region.Networks[0].String() != ipnet.String()  {
			t.Error("Failed to parse region map")
		}
	}
	if region, ok := rm["us"]; ok {
		_, ipnet, _ := net.ParseCIDR("10.0.3.0/14")
		if region.Region != "us" || region.Networks[0].String() != ipnet.String()  {
			t.Error("Failed to parse region map")
		}
		_, ipnet, _ = net.ParseCIDR("1.1.1.1/32")
		if region.Region != "us" || region.Networks[1].String() != ipnet.String()  {
			t.Error("Failed to parse region map")
		}
	}
}

func TestIsClientRegion(t *testing.T) {
	_, rm := NewRegionMap("test_data/clientMap.yaml")
	if !rm["il"].IsClientInRegion("192.168.1.32") {
		t.Error("Failed to validate ip address")
	}
	if rm["il"].IsClientInRegion("192.168.3.32") {
		t.Error("Failed to validate ip address")
	}
	if !rm["us"].IsClientInRegion("1.1.1.1") {
		t.Error("Failed to validate ip address")
	}
	if !rm["us"].IsClientInRegion("10.1.32.3") {
		t.Error("Failed to validate ip address")
	}
}

func TestGetRegion(t *testing.T) {
	_, rm := NewRegionMap("test_data/clientMap.yaml")
	if rm.GetRegion("192.168.1.32") != "il" {
		t.Error("Failed to find region")
	}
	if rm.GetRegion("1.1.1.1") != "us" {
		t.Error("Failed to find region")
	}
	if rm.GetRegion("10.1.32.3") != "us" {
		t.Error("Failed to find region")
	}
	if rm.GetRegion("10.7.32.3") != "" {
		t.Error("Failed to find region")
	}
	if rm.GetRegion("1.1.1.2") != "" {
		t.Error("Failed to find region")
	}
}