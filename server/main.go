package main

import (
	"kayena/server/config"
	"kayena/server/database"
	"kayena/server/http/routes"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http"

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
	err = database.SeedDb(service)
	if err != nil {
		logger.Fatal(err)
	}

	router := routes.NewRouter(service)
	server := http.Server{
		Handler: golog.Log(router),
		Addr:    c.Port,
	}

	logger.Infof("Server listening on port %s", c.Port)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
