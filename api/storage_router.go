package api

import (
	"net/http"

	"github.com/dzjyyds666/mediaStorage/internal/handler"
	"github.com/dzjyyds666/vortex/v2"
)

func PrepareRouters(login *handler.LoginHandler, file *handler.FileHandler) []*vortex.VortexHttpRouter {
	return []*vortex.VortexHttpRouter{
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/login", login.HandleLogin, "登录接口"),

		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/deport/create", file.HandleDeportCreate, "创建 depot"),
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/box/create", file.HandleBoxCreate, "创建 box"),

		vortex.AppendHttpRouter([]string{http.MethodPost, http.MethodGet, http.MethodHead}, "/media/file/:fid", file.HandleFile, "查看文件"),
		vortex.AppendHttpRouter([]string{http.MethodGet}, "/media/file/info/:fid", file.HandleFileInfo, "查看文件"),
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/upload/apply", file.HandleApplyUpload, "申请上传"),
		vortex.AppendHttpRouter([]string{http.MethodPost}, "/media/upload/single/:fid", file.HandleSingleUpload, "单文件上传"),
	}
}
