package main

import (
	"kayena/server/config"
	"kayena/server/database"
	repository "kayena/server/database/generated"
	"kayena/server/http/routes"
	"kayena/server/utils"
	"net/http"

	"github.com/K44Z/golog"
)

func main() {
	var err error

	logger := utils.InitLogger()
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(err)
	}

	db, err := database.ConnectDb(cfg)
	if err != nil {
		logger.Fatal("Connection to Db failed: ", err)
	}

	repo := repository.New(db)

	router := routes.NewRouter()
	server := http.Server{
		Handler: golog.Log(router),
		Addr:    cfg.Port,
	}

	logger.Infof("Server listening on port %s", cfg.Port)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
