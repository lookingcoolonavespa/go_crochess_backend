package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/database"
	delivery_ws_game "github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/delivery/ws"
	repository_game "github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/repository"
	usecase_game "github.com/lookingcoolonavespa/go_crochess_backend/src/services/game/usecase"
	delivery_ws_gameseeks "github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/delivery/ws"
	repository_gameseeks "github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/repository"
	usecase_gameseeks "github.com/lookingcoolonavespa/go_crochess_backend/src/services/gameseeks/usecase"

	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/spf13/viper"
)

func Run() {
	initConfig()

	db, err := initDB()
	if err != nil {
		log.Fatalf("%s: %v", "Error on connect to database", err)
	}

	initHandlers(db)
}

func initConfig() {
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")
	viper.SetConfigName(".config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}
}

func initDB() (*sql.DB, error) {
	dbConnector := database.DatabaseConnector{
		Host:     viper.GetString("database.host"),
		Username: viper.GetString("database.Username"),
		Password: viper.GetString("database.password"),
		Port:     viper.GetInt("database.port"),
		DBName:   viper.GetString("database.name"),
	}

	db, err := dbConnector.Connect()
	if err != nil {
		return nil, err
	}

	// err = migrations.Up(db)
	// if err != nil {
	// 	log.Fatalf("error on migratre schema: %v", err)
	// }

	return db, nil
}

func initHandlers(db *sql.DB) {
	gameseeksTopic, err := domain_websocket.NewTopic("gameseeks")
	if err != nil {
		log.Printf("error instantiating gameseeks topic: %v", err)
		return
	}
	gameseeksRepo := repository_gameseeks.NewGameseeksRepo(db)
	gameRepo := repository_game.NewGameRepo(db)
	gameseeksUseCase := usecase_gameseeks.NewGameseeksUseCase(db, gameRepo)
	gameseeksHandler := delivery_ws_gameseeks.NewGameseeksHandler(gameseeksRepo, gameseeksUseCase, gameseeksTopic)
	gameseeksTopic.RegisterEvent(domain_websocket.SubscribeEvent, gameseeksHandler.HandlerGetGameseeksList)
	gameseeksTopic.RegisterEvent(domain_websocket.InsertEvent, gameseeksHandler.HandleGameseekInsert)
	gameseeksTopic.RegisterEvent(domain_websocket.UnsubscribeEvent, gameseeksHandler.HandlerOnUnsubscribe)

	gameTopic, err := domain_websocket.NewTopic("game/id")
	if err != nil {
		log.Printf("error instantiating gameseeks topic: %v", err)
		return
	}
	gameUseCase := usecase_game.NewGameUseCase(db, gameRepo)
	gameHandler := delivery_ws_game.NewGameHandler(gameTopic.(domain_websocket.TopicWithParam), gameUseCase)
	gameTopic.RegisterEvent(domain_websocket.SubscribeEvent, gameHandler.HandlerOnSubscribe)

	webSocketRouter, err := domain_websocket.NewWebSocketRouter()
	if err != nil {
		log.Printf("error instantiating web socket router: %v", err)
		return
	}
	webSocketRouter.PushNewRoute(gameTopic)
	webSocketRouter.PushNewRoute(gameseeksTopic)

	webSocketServer := domain_websocket.NewWebSocketServer(webSocketRouter, gameseeksRepo)

	http.HandleFunc("/ws", webSocketServer.HandleWS)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", viper.GetInt("app.port")),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	webSocketServer.Close()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

}
