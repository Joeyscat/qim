package ipregion

import "github.com/lionsoul2014/ip2region/binding/golang/ip2region"

type IPInfo struct {
	Country string
	Region  string
	City    string
	ISP     string
}

type IPRegion interface {
	Search(ip string) (*IPInfo, error)
}

type IP2Region struct {
	region *ip2region.Ip2Region
}

func NewIP2Region(path string) (*IP2Region, error) {
	if path == "" {
		path = "ip2region.db"
	}

	region, err := ip2region.New(path)
	if err != nil {
		return nil, err
	}

	return &IP2Region{region: region}, nil
}

func (r *IP2Region) Search(ip string) (*IPInfo, error) {
	data, err := r.region.MemorySearch(ip)
	if err != nil {
		return nil, err
	}

	return &IPInfo{
		Country: data.Country,
		Region:  data.Region,
		City:    data.City,
		ISP:     data.ISP,
	}, nil
}
