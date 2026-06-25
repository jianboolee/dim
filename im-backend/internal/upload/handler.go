package upload

import (
	"errors"
	"net/http"

	"d-im/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// UploadImage 单图上传，表单字段 file
func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	res, err := h.svc.UploadImage(c.Request.Context(), file)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	response.Success(c, "success", res)
}

// UploadImages 批量图片上传，表单字段 files；兼容单文件字段 file。
func (h *Handler) UploadImages(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		writeUploadError(c, ErrNoFile)
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}

	res, err := h.svc.UploadImages(c.Request.Context(), files)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	response.Success(c, "success", res)
}

func writeUploadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNoFile):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrStorageUnavailable):
		response.Error(c, http.StatusServiceUnavailable, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrInvalidImage):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrInvalidImageURL):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrTooManyFiles):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrFileTooLarge):
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c, err.Error())
	}
}
