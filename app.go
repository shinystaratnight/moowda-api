package main

import (
	"database/sql"
	"flag"
	"fmt"
	"moowda/models"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-playground/validator"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"

	"moowda/app"
	apiErrors "moowda/errors"
	"moowda/web"
)

var addr = flag.String("addr", ":8080", "http service address")

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if err == sql.ErrNoRows {
		err = apiErrors.NotFound("the requested resource")
		return
	}

	switch err.(type) {
	case validation.Errors:
		err = apiErrors.InvalidData(err.(validation.Errors))
	}

	if gorm.IsRecordNotFoundError(err) {
		err = apiErrors.NotFound("the requested resource")
	}

	if _, ok := err.(*apiErrors.APIError); ok {
		httpError := err.(*apiErrors.APIError)
		if err := c.JSON(httpError.StatusCode(), httpError); err != nil {
			c.Logger().Error(err)
		}
		return
	}

	if _, ok := err.(*echo.HTTPError); ok {
		httpError := err.(*echo.HTTPError)
		if err := c.JSON(httpError.Code, httpError); err != nil {
			c.Logger().Error(err)
		}
		return
	}

	c.Logger().Error(c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()}))
}

func run() {
	flag.Parse()

	if err := app.LoadConfig("./config"); err != nil {
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	// load error messages
	if err := apiErrors.LoadMessages(app.Config.ErrorFile); err != nil {
		panic(fmt.Errorf("failed to read the error message file: %s", err))
	}

	e := echo.New()
	e.Use(middleware.BodyLimit("5M"))

	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/media/uploads", app.Config.UploadPath)

	corsConfig := middleware.DefaultCORSConfig
	corsConfig.AllowCredentials = true
	e.Use(middleware.CORSWithConfig(corsConfig))

	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Debug = true

	// =========================================================================
	// Start Postgres
	db, err := gorm.Open("postgres", app.Config.DSN)
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(models.Image{})

	db.DB().SetMaxOpenConns(50)
	db.DB().SetMaxIdleConns(50)

	// Logging
	db.LogMode(true)
	defer db.Close()

	web.AddRoutes(e, db)

	e.Logger.Fatal(e.Start(*addr))
}
