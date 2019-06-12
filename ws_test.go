package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"net/http/httptest"

	"moowda/app"
	"moowda/web"
)

func TestWebsocketAuth(t *testing.T) {
	if err := app.LoadConfig("./config"); err != nil {
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	e := echo.New()

	db, err := gorm.Open("postgres", app.Config.DSN)
	if err != nil {
		panic(err)
	}

	web.AddRoutes(e, db)
	server := httptest.NewServer(http.Handler(e))
	defer server.Close()

	wsURL := fmt.Sprintf("ws://%s/ws/topics/events?token=%s",
		strings.TrimPrefix(server.URL, "http://"),
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjA1MzkwMDUsInVzZXJJRCI6MX0.aOT98IgqF5nDjD89QCK8ydwkx07YxjbMLptOHMRM5pE")

	header := http.Header{}
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
	}
	defer ws.Close()
}
