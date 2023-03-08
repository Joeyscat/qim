package qim

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var channelTotalGauge = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "qim",
		Name: "channel_total",
		Help: "网关并发数",
	},
	[]string{"serviceID", "serviceName"},
)