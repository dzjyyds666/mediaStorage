package server

import (
	"net/http"

	"github.com/dzjyyds666/vortex/v2"
)

func PrepareRouters(h *StorageServer) []*vortex.VortexHttpRouter {

	return []*vortex.VortexHttpRouter{
		vortex.AppendHttpRouter([]string{http.MethodGet}, "/test", h.HandleRouterTest, "测试路由"),
	}
}
