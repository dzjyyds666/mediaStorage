package locale

var V = "{\"code_for_bad_request.en-us\":\"bad request\",\"code_for_bad_request.zh-cn\":\"错误请求\",\"code_for_box_not_exists.en-us\":\"box not exists\",\"code_for_box_not_exists.zh-cn\":\"box不存在\",\"code_for_file_exists.en-us\":\"file exists\",\"code_for_file_exists.zh-cn\":\"文件已存在\",\"code_for_file_no_prepare_info.en-us\":\"file no prepare info\",\"code_for_file_no_prepare_info.zh-cn\":\"文件未初始化上传\",\"code_for_file_not_exists.en-us\":\"file not exists\",\"code_for_file_not_exists.zh-cn\":\"文件不存在\",\"code_for_internal_error.en-us\":\"internal error\",\"code_for_internal_error.zh-cn\":\"服务器内部错误\",\"code_for_permission_deny.en-us\":\"permission deny\",\"code_for_permission_deny.zh-cn\":\"权限不足\"}"

var K = struct {
	CODE_FOR_PERMISSION_DENY string
	CODE_FOR_FILE_EXISTS string
	CODE_FOR_FILE_NOT_EXISTS string
	CODE_FOR_FILE_NO_PREPARE_INFO string
	CODE_FOR_BOX_NOT_EXISTS string
	CODE_FOR_BAD_REQUEST string
	CODE_FOR_INTERNAL_ERROR string
} {
	CODE_FOR_FILE_NOT_EXISTS: "code_for_file_not_exists",
	CODE_FOR_FILE_NO_PREPARE_INFO: "code_for_file_no_prepare_info",
	CODE_FOR_BOX_NOT_EXISTS: "code_for_box_not_exists",
	CODE_FOR_BAD_REQUEST: "code_for_bad_request",
	CODE_FOR_INTERNAL_ERROR: "code_for_internal_error",
	CODE_FOR_PERMISSION_DENY: "code_for_permission_deny",
	CODE_FOR_FILE_EXISTS: "code_for_file_exists",
}
