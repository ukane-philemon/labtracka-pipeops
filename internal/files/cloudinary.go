package files

import (
	"context"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryClient struct {
	*cloudinary.Cloudinary
}

func NewCloudinaryClient() (*CloudinaryClient, error) {
	cld, err := cloudinary.New()
	if err != nil {
		return nil, err
	}

	cld.Config.URL.Secure = true
	return &CloudinaryClient{
		Cloudinary: cld,
	}, nil
}

// Upload sends the file to the file database.
func (cc *CloudinaryClient) UploadFile(ctx context.Context, dir, fileName string, file io.Reader) (string, error) {
	// Upload the image. Set the asset's public ID and allow overwriting the
	// asset with new versions.
	yes := true
	resp, err := cc.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:                       fileName,
		UseFilename:                    &yes,
		UniqueFilename:                 &yes,
		UseFilenameAsDisplayName:       &yes,
		Folder:                         dir,
		UseAssetFolderAsPublicIDPrefix: &yes,
		Overwrite:                      api.Bool(true),
		Transformation:                 "",
		Format:                         "",
		OnSuccess:                      "",
		Callback:                       "",
		NotificationURL:                "",
	})
	if err != nil {
		return "", nil
	}

	return resp.SecureURL, nil
}
