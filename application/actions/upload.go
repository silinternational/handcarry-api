package actions

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/models"
)

// fileFieldName is the multipart field name for the file upload.
const fileFieldName = "file"

// UploadResponse is a JSON response for the /upload endpoint
type UploadResponse struct {
	Error       *domain.AppError `json:"Error,omitempty"`
	Name        string           `json:"filename,omitempty"`
	UUID        string           `json:"id,omitempty"`
	URL         string           `json:"url,omitempty"`
	ContentType string           `json:"content_type,omitempty"`
	Size        int              `json:"size,omitempty"`
}

// uploadHandler responds to POST requests at /upload
func uploadHandler(c buffalo.Context) error {
	f, err := c.File(fileFieldName)
	if err != nil {
		domain.Error(c, fmt.Sprintf("error getting uploaded file from context ... %v", err))
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{
				Code: http.StatusInternalServerError,
				Key:  domain.ErrorReceivingFile,
			},
		}))
	}

	if f.Size > int64(domain.MaxFileSize) {
		domain.Error(c, fmt.Sprintf("file upload size (%v) greater than max (%v)", f.Size, domain.MaxFileSize))
		return c.Render(http.StatusBadRequest, render.JSON(domain.AppError{
			Code: http.StatusBadRequest,
			Key:  domain.ErrorStoreFileTooLarge,
		}))
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		domain.Error(c, fmt.Sprintf("error reading uploaded file ... %v", err))
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{
				Code: http.StatusInternalServerError,
				Key:  domain.ErrorUnableToReadFile,
			},
		}))
	}

	fileObject := models.File{
		Name:    f.Filename,
		Content: content,
	}
	if fErr := fileObject.Store(models.Tx(c)); fErr != nil {
		domain.Error(c, fmt.Sprintf("error storing uploaded file ... %v", fErr))
		return c.Render(fErr.HttpStatus, render.JSON(domain.AppError{
			Code: fErr.HttpStatus,
			Key:  fErr.ErrorCode,
		}))
	}

	resp := UploadResponse{
		Name:        fileObject.Name,
		UUID:        fileObject.UUID.String(),
		URL:         fileObject.URL,
		ContentType: fileObject.ContentType,
		Size:        fileObject.Size,
	}

	return c.Render(200, render.JSON(resp))
}
