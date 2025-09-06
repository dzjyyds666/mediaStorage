package pkg

import (
	"github.com/dzjyyds666/mediaStorage/locale"
	"github.com/dzjyyds666/vortex/v2"
)

var SubStatusCodes = struct {
	BadRequest     vortex.SubCode // 400 错误请求
	InternalError  vortex.SubCode // 500 服务器内部错误
	PermissionDeny vortex.SubCode // 403 权限不足

	FileExist         vortex.SubCode // 20001
	FileNotExist      vortex.SubCode // 20404
	NoPrepareFileInfo vortex.SubCode // 20002

	BoxNotExist vortex.SubCode // 30404
}{

	BadRequest:     vortex.SubCode{SubCode: 400, I18nKey: locale.K.CODE_FOR_BAD_REQUEST},
	InternalError:  vortex.SubCode{SubCode: 500, I18nKey: locale.K.CODE_FOR_INTERNAL_ERROR},
	PermissionDeny: vortex.SubCode{SubCode: 403, I18nKey: locale.K.CODE_FOR_PERMISSION_DENY},

	FileExist:         vortex.SubCode{SubCode: 20001, I18nKey: locale.K.CODE_FOR_FILE_EXISTS},
	FileNotExist:      vortex.SubCode{SubCode: 20404, I18nKey: locale.K.CODE_FOR_FILE_NOT_EXISTS},
	NoPrepareFileInfo: vortex.SubCode{SubCode: 20002, I18nKey: locale.K.CODE_FOR_FILE_NO_PREPARE_INFO},

	BoxNotExist: vortex.SubCode{SubCode: 30404, I18nKey: locale.K.CODE_FOR_BOX_NOT_EXISTS},
}
