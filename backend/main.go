package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/ryo02puyopuyo/sudoku_online/backend/api"
	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
	"github.com/ryo02puyopuyo/sudoku_online/backend/hub"
	"github.com/ryo02puyopuyo/sudoku_online/backend/middleware"
)

func main() {
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}

	gameInstance := game.NewGame()
	hubInstance := hub.NewHub(gameInstance)
	apiHandlers := &api.API{DB: dbConn}
	authMiddleware := &middleware.Auth{DB: dbConn}

	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/register", apiHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", apiHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/test", apiHandlers.TestHandler).Methods("POST", "OPTIONS")

	//tuika
	apiRouter.Handle("/me", authMiddleware.Optional(http.HandlerFunc(apiHandlers.MeHandler))).Methods("GET", "OPTIONS")

	r.Handle("/ws", authMiddleware.Optional(http.HandlerFunc(hubInstance.ServeWs)))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true, // Cookieを送受信するために必須
	})
	handler := c.Handler(r)

	log.Println("サーバーがポート8080で起動しました")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
