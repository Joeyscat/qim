package conf

import (
	"encoding/json"
	"os"
)

type Zone struct {
	ID     string
	Weight int
}

type Route struct {
	RouteBy   string
	Zones     []Zone
	Whitelist map[string]string
	Slots     []int
}

func ReadRoute(path string) (*Route, error) {
	var conf struct {
		RouteBy   string `json:"route_by,omitempty"`
		Zones     []Zone `json:"zones,omitempty"`
		Whitelist []struct {
			Key   string `json:"key,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"whitelist,omitempty"`
	}

	bts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bts, &conf)
	if err != nil {
		return nil, err
	}

	var rt = Route{
		RouteBy:   conf.RouteBy,
		Zones:     conf.Zones,
		Whitelist: make(map[string]string, len(conf.Whitelist)),
		Slots:     make([]int, 0),
	}

	// build slots
	for i, zone := range conf.Zones {
		shard := make([]int, zone.Weight)
		for j := 0; j < zone.Weight; j++ {
			shard[j] = i
		}
		rt.Slots = append(rt.Slots, shard...)
	}
	for _, wl := range conf.Whitelist {
		rt.Whitelist[wl.Key] = wl.Value
	}

	return &rt, nil
}
