package serv

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var messageInTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "qim",
	Name:      "message_in_total",
	Help:      "网关接收消息总数",
}, []string{"serviceId", "serviceName", "command"})

var messageInFlowBytes = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "qim",
	Name:      "message_in_flow_bytes",
	Help:      "网关接收消息字节数",
}, []string{"serviceId", "serviceName", "command"})

var serverNotFoundErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "qim",
	Name:      "server_not_found_error_total",
	Help:      "在zone分区中查找服务失败次数",
}, []string{"zone"})
