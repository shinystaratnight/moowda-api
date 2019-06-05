package main

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"moowda/app"

	"moowda/web"
)

var addr = flag.String("addr", ":8080", "http service address")

func run() {
	flag.Parse()

	e := echo.New()

	corsConfig := middleware.DefaultCORSConfig
	corsConfig.AllowCredentials = true
	e.Use(middleware.CORSWithConfig(corsConfig))

	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Debug = true

	if err := app.LoadConfig("./config"); err != nil {
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	// =========================================================================
	// Start Postgres
	db, err := gorm.Open("postgres", app.Config.DSN)
	if err != nil {
		panic(err)
	}

	// Logging
	db.LogMode(true)
	defer db.Close()

	web.AddRoutes(e, db)

	//web.RunHub(e)

	e.Logger.Fatal(e.Start(*addr))
}
