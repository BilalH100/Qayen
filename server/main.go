package main

import (
	"kayena/server/database"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	PORT := os.Getenv("PORT")
	err = database.ConnectDb()
	if err != nil {
		log.Fatal("Error connecting to the database :", err)
	}
	log.Println("Database connected ✅")
	app := fiber.New()
	app.Use(logger.New())
	err = app.Listen(PORT)
	if err != nil {
		panic(err)
	}
}
