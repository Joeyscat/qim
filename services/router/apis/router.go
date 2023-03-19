package apis

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/services/router/conf"
	"github.com/joeyscat/qim/services/router/ipregion"
	"github.com/joeyscat/qim/wire"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

const DefaultLocation = "CN"

type RouterApi struct {
	Naming   naming.Naming
	IPRegion ipregion.IPRegion
	Config   conf.Router
	Lg       *zap.Logger
}

type LookupResp struct {
	UTC      int64    `json:"utc"`
	Location string   `json:"location"`
	Domains  []string `json:"domains"`
}

func (r *RouterApi) Lookup(c iris.Context) {
	ip := qim.FromRequest(c.Request())
	token := c.Params().Get("token")

	var location conf.Country
	ipinfo, err := r.IPRegion.Search(ip)
	if err != nil || ipinfo.Country == "0" {
		location = DefaultLocation
	} else {
		location = conf.Country(ipinfo.Country)
	}

	regionID, ok := r.Config.Mapping[location]
	if !ok {
		c.StopWithError(iris.StatusForbidden, err)
		return
	}

	region, ok := r.Config.Regions[regionID]
	if !ok {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	idc := selectIdc(token, region)

	gateways, err := r.Naming.Find(wire.SNWGateway, fmt.Sprintf("IDC:%s", idc.ID))
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	hits := selectGateways(token, gateways, 3)
	domains := make([]string, len(hits))
	for i, hit := range hits {
		domains[i] = hit.GetMeta()["domain"]
	}

	r.Lg.Info("lookup", zap.Any("country", location), zap.String("regionID", regionID),
		zap.String("idc", idc.ID), zap.Any("domains", domains))

	_ = c.JSON(LookupResp{
		UTC:      time.Now().Unix(),
		Location: string(location),
		Domains:  domains,
	})
}

func selectIdc(token string, region *conf.Region) *conf.IDC {
	slot := hashcode(token) % len(region.Slots)
	i := region.Slots[slot]
	return &region.Idcs[i]
}

func selectGateways(token string, gateways []qim.ServiceRegistration, num int) []qim.ServiceRegistration {
	if len(gateways) <= num {
		return gateways
	}

	slots := make([]int, len(gateways)*10)
	for i := 0; i < len(gateways); i++ {
		for j := 0; j < 10; j++ {
			slots[i*10+j] = i
		}
	}

	slot := hashcode(token) % len(slots)
	i := slots[slot]

	hits := make([]qim.ServiceRegistration, 0, num)
	for len(hits) < num {
		hits = append(hits, gateways[i])
		i++
		if i >= len(gateways) {
			i = 0
		}
	}

	return hits
}

func hashcode(s string) int {
	hash32 := crc32.NewIEEE()
	_, _ = hash32.Write([]byte(s))
	return int(hash32.Sum32())
}
