package main

import (
	"kayena/server/config"
	"kayena/server/database"
	"kayena/server/http/routes"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http"
	"strings"

	"github.com/K44Z/golog"
)

func main() {
	var err error

	logger := utils.InitLogger()
	c, err := config.Load()
	if err != nil {
		logger.Fatal(err)
	}

	db, err := database.ConnectDb(c)
	if err != nil {
		logger.Fatal("Connection to Db failed: ", err)
	}
	service := services.NewService(db)
	err = database.SeedDb(service, logger)
	if err != nil {
		logger.Fatal(err)
	}

	router := routes.NewRouter(service)
	loggedRouter := golog.Log(router)

	// golog.Log buffers the whole response before writing it out, which
	// works fine for normal request/response endpoints but breaks Server-
	// Sent Events: the SSE handlers (/alerts/{id}/stream and
	// /pharmacies/{id}/alerts/stream) are meant to stay open and stream
	// data as it happens, so nothing ever reaches the client through a
	// buffering wrapper. Route those two endpoints straight to the router,
	// unwrapped, and keep logging for everything else.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/stream") {
			router.ServeHTTP(w, r)
			return
		}
		loggedRouter.ServeHTTP(w, r)
	})

	server := http.Server{
		Handler: handler,
		Addr:    c.Port,
	}

	logger.Infof("Server listening on port %s", c.Port)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
