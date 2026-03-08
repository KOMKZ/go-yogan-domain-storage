package errors

import (
	"net/http"

	"github.com/KOMKZ/go-yogan-framework/errcode"
)

const ModuleStorage = 21

var (
	ErrDatabaseError = errcode.Register(errcode.New(
		ModuleStorage, 1001, "storage",
		"error.storage.database_error", "数据库操作失败",
		http.StatusInternalServerError,
	))

	ErrNotFound = errcode.Register(errcode.New(
		ModuleStorage, 1002, "storage",
		"error.storage.not_found", "文件不存在",
		http.StatusNotFound,
	))

	ErrBadRequest = errcode.Register(errcode.New(
		ModuleStorage, 1003, "storage",
		"error.storage.bad_request", "请求参数错误",
		http.StatusBadRequest,
	))

	ErrUploadFailed = errcode.Register(errcode.New(
		ModuleStorage, 1004, "storage",
		"error.storage.upload_failed", "文件上传失败",
		http.StatusInternalServerError,
	))

	ErrReadFailed = errcode.Register(errcode.New(
		ModuleStorage, 1005, "storage",
		"error.storage.read_failed", "读取文件失败",
		http.StatusInternalServerError,
	))
)
