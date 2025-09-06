package api

import (
	"net/http"

	"github.com/dzjyyds666/vortex/v2"
)

func PrepareRouters(h *StorageServer) []*vortex.VortexHttpRouter {

	return []*vortex.VortexHttpRouter{
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/login", h.HandleLogin, "登录接口"),

		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/deport/create", h.HandleDeportCreate, "创建 depot"),
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/box/create", h.HandleBoxCreate, "创建 box"),

		vortex.AppendHttpRouter([]string{http.MethodPost, http.MethodGet, http.MethodHead}, "/media/file/:fid", h.HandleFile, "查看文件"),
		vortex.AppendHttpRouter([]string{http.MethodGet}, "/media/file/info/:fid", h.HandleFileInfo, "查看文件"),
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/upload/apply", h.HandleApplyUpload, "申请上传"),
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/upload/single/:fid", h.HandleSingleUpload, "单文件上传"),
	}
}
