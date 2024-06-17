package files

import (
	"context"
	"io"
)

type FileUploader interface {
	// UploadFile sends the file to the file database.
	UploadFile(ctx context.Context, dir, fileName string, file io.Reader) (string, error)
}

func IsFileTypeSupported(fileType string) bool {
	switch fileType {
	case "image/png", "image/jpeg", "image/jpg", "application/pdf":
		return true
	}
	return false
}

func IsSupportedImageFile(fileType string) bool {
	return IsFileTypeSupported(fileType) && fileType != "application/pdf"
}
