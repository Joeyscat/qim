package handler

import "github.com/joeyscat/qim"

type loginHandler struct {
}

func NewLoginHandler() *loginHandler {
	return &loginHandler{}
}

func (h *loginHandler) DoSysLogin(ctx qim.Context) {
	panic("unimplemented")
}

func (h *loginHandler) DoSysLogout(ctx qim.Context) {
	panic("unimplemented")
}
