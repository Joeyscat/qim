package middleware

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/logger"
	"github.com/joeyscat/qim/wire/pkt"
	"go.uber.org/zap"
)

func Recover() qim.HandlerFunc {
	return func(ctx qim.Context) {
		defer func() {
			if err := recover(); err != nil {
				var callers []string
				for i := 1; ; i++ {
					_, file, line, got := runtime.Caller(i)
					if !got {
						break
					}
					callers = append(callers, fmt.Sprintf("%s:%d", file, line))
				}
				logger.L.Error(strings.Join(callers,"\n"), zap.Any("error", err),
					zap.String("ChannelID", ctx.Header().GetChannelId()),
					zap.String("Command", ctx.Header().GetCommand()),
					zap.Uint32("Seq", ctx.Header().GetSequence()),
				)

				_ = ctx.Resp(pkt.Status_SystemException, &pkt.ErrorResp{Message: "SystemException"})
			}
		}()
	}
}
