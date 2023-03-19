package conf

import (
	"encoding/json"
	"os"
)

type IDC struct {
	ID     string
	Weight int
}

type Region struct {
	ID    string
	Idcs  []IDC
	Slots []byte
}

type Country string

type Mapping struct {
	Region   string
	Location []string
}

type Router struct {
	Mapping map[Country]string
	Regions map[string]*Region
}

func LoadMapping(path string) (map[Country]string, error) {
	bts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mps []Mapping
	err = json.Unmarshal(bts, &mps)
	if err != nil {
		return nil, err
	}
	mp := make(map[Country]string)
	for _, m := range mps {
		for _, l := range m.Location {
			mp[Country(l)] = m.Region
		}
	}

	return mp, nil
}

func LoadRegions(path string) (map[string]*Region, error) {
	bts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var regions []*Region
	err = json.Unmarshal(bts, &regions)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*Region)
	for _, region := range regions {
		res[region.ID] = region
		for _, idc := range region.Idcs {

			shard := make([]byte, idc.Weight)
			for i := 0; i < idc.Weight; i++ {
				shard[i] = byte(i)
			}

			region.Slots = append(region.Slots, shard...)
		}
	}

	return res, nil
}
