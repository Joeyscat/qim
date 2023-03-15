package serv

import (
	"hash/crc32"
	"math/rand"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/container"
	"github.com/joeyscat/qim/services/gateway/conf"
	"github.com/joeyscat/qim/wire/pkt"
	"go.uber.org/zap"
)

type RouteSelector struct {
	route *conf.Route
	lg    *zap.Logger
}

func NewRouteSelector(configPath string, lg *zap.Logger) (*RouteSelector, error) {
	route, err := conf.ReadRoute(configPath)
	if err != nil {
		return nil, err
	}
	return &RouteSelector{
		route: route,
		lg:    lg,
	}, nil
}

var _ container.Selector = (*RouteSelector)(nil)

// Lookup implements container.Selector
func (s *RouteSelector) Lookup(header *pkt.Header, srvs []qim.Service) string {
	// read meta from header
	app, _ := pkt.FindMeta(header.Meta, MetaKeyApp)
	accout, _ := pkt.FindMeta(header.Meta, MetaKeyAccount)
	if app == nil || accout == nil {
		ri := rand.Intn(len(srvs))
		return srvs[ri].ServiceID()
	}

	log := s.lg.With(zap.String("app", app.(string)), zap.String("account", accout.(string)))

	zone, ok := s.route.Whitelist[app.(string)]
	if !ok {
		var key string
		switch s.route.RouteBy {
		case MetaKeyApp:
			key = app.(string)
		case MetaKeyAccount:
			key = accout.(string)
		default:
			key = accout.(string)
		}

		slot := hashcode(key) % len(s.route.Slots)
		i := s.route.Slots[slot]
		zone = s.route.Zones[i].ID
	} else {
		log.Info("hit a zone in whitelist", zap.String("zone", zone))
	}

	zoneSrvs := filterSrvs(srvs, zone)
	if len(zoneSrvs) == 0 {
		serverNotFoundErrorTotal.WithLabelValues(zone).Inc()
		log.Warn("select a random service from all due to no service found in zone", zap.String("zone", zone))
		ri := rand.Intn(len(srvs))
		return srvs[ri].ServiceID()
	}

	srv := selectSsrvs(zoneSrvs, accout.(string))
	return srv.ServiceID()
}

func filterSrvs(srvs []qim.Service, zone string) []qim.Service {
	var res = make([]qim.Service, 0, len(srvs))
	for _, srv := range srvs {
		if zone == srv.GetMeta()["zone"] {
			res = append(res, srv)
		}
	}
	return res
}

func selectSsrvs(srvs []qim.Service, account string) qim.Service {
	slots := make([]int, 0, len(srvs)*10)
	for i := range srvs {
		for j := 0; j < 10; j++ {
			slots = append(slots, i)
		}
	}
	slot := hashcode(account) % len(slots)
	return srvs[slots[slot]]
}

func hashcode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}
