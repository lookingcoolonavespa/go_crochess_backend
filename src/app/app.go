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
	// "github.com/lookingcoolonavespa/go_crochess_backend/database/migrations"
	domain_websocket "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket"
	"github.com/spf13/viper"
)

func Run() {
	initConfig()

	_, err := initDB()
	if err != nil {
		log.Fatalf("%s: %v", "Error on connect to database", err)
	}

	initHandlers()
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

func initHandlers() {
	webSocketRouter, err := domain_websocket.NewWebSocketRouter()
	if err != nil {
		log.Printf("error instantiating web socket router: %v", err)
		return
	}
	webSocketServer := domain_websocket.NewWebSocketServer(webSocketRouter)
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

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

}
