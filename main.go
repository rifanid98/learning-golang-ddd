package main

import (
	"learning-golang-ddd/infrastructure/auth"
	"learning-golang-ddd/infrastructure/persistence"
	"learning-golang-ddd/interface/fileupload"
	"learning-golang-ddd/interface/handler"
	middleware "learning-golang-ddd/interface/middeware"

	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// to load our environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("no env gotten")
	}
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// redis details
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	services, err := persistence.NewRepositories(
		dbUser,
		dbPassword,
		dbPort,
		dbHost,
		dbName,
	)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.Automigrate()

	redisService, err := auth.NewRedisDB(redisHost, redisPort, redisPassword)
	if err != nil {
		log.Fatal(err)
	}

	ti := auth.NewToken()
	fileUpload := fileupload.NewFileUpload()

	users := handler.NewUsersHandler(services.User, redisService.Auth, ti)
	foods := handler.NewFoodHandler(services.Food, services.User, fileUpload, redisService.Auth, ti)
	auth := handler.NewAuthenticateHandler(services.User, redisService.Auth, ti)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware()) // For CORS

	//user routes
	r.POST("/users", users.SaveUser)
	r.GET("/users", users.GetUsers)
	r.GET("/users/:user_id", users.GetUser)

	//post routes
	r.POST("/food", middleware.AuthMiddleware(), middleware.MaxSizeAllowed(8192000), foods.SaveFood)
	r.PUT("/food/:food_id", middleware.AuthMiddleware(), middleware.MaxSizeAllowed(8192000), foods.UpdateFood)
	r.GET("/food/:food_id", foods.GetFoodAndCreator)
	r.DELETE("/food/:food_id", middleware.AuthMiddleware(), foods.DeleteFood)
	r.GET("/food", foods.GetAllFood)

	//authentication routes
	r.POST("/auth/login", auth.Login)
	r.POST("/auth/logout", auth.Logout)
	r.POST("/auth/refresh", auth.Refresh)

	//Starting the application
	appPort := os.Getenv("APP_PORT") //using heroku host
	log.Print(appPort)
	if appPort == "" {
		appPort = "8080" //localhost
	}
	log.Fatal(r.Run(":" + appPort))
}
