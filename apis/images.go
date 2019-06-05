package apis

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"moowda/storage"
	"net/http"
)

type ImagesAPI struct {
	db          *gorm.DB
	fileStorage storage.FileStorage
}

func NewImagesAPI(db *gorm.DB, storage storage.FileStorage) *ImagesAPI {
	return &ImagesAPI{db: db, fileStorage: storage}
}

func (r *ImagesAPI) Upload(c echo.Context) error {
	sourceFile, err := c.FormFile("file")
	if err != nil {
		return errors.Wrap(err, "get form file")
	}

	file, err := sourceFile.Open()
	if err != nil {
		return errors.Wrap(err, "open form file")
	}

	url, err := r.fileStorage.Store(c, sourceFile.Filename, file)
	if err != nil {
		return errors.Wrap(err, "store file")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url": url,
	})
}
