package apis

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"moowda/models"
	"moowda/storage"
)

type ImagesAPI struct {
	db          *gorm.DB
	fileStorage storage.FileStorage
}

func NewImagesAPI(db *gorm.DB, storage storage.FileStorage) *ImagesAPI {
	return &ImagesAPI{db: db, fileStorage: storage}
}

func (r *ImagesAPI) Upload(c echo.Context) error {
	user := c.Get("user").(*models.User)

	sourceFile, err := c.FormFile("file")
	if err != nil {
		return errors.Wrap(err, "get form file")
	}

	file, err := sourceFile.Open()
	if err != nil {
		return errors.Wrap(err, "open form file")
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return errors.Wrap(err, "parse image")
	}

	file.Seek(0, 0)

	url, err := r.fileStorage.Store(c, sourceFile.Filename, file)
	if err != nil {
		return errors.Wrap(err, "store file")
	}

	image := models.Image{
		UserID: user.ID,
		URL:    url,
		Width:  config.Width,
		Height: config.Height,
	}

	if err := r.db.Create(&image).Error; err != nil {
		return errors.Wrap(err, "create image")
	}

	return c.JSON(http.StatusOK, image)
}
