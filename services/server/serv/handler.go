package serv

import (
	"github.com/joeyscat/qim"
	"go.uber.org/zap"
)

type ServHandler struct {
	r     *qim.Router
	cache qim.SessionStorage
	Lg    *zap.Logger
}
