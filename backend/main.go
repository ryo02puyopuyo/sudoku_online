package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/ryo02puyopuyo/sudoku_online/backend/api"
	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"github.com/ryo02puyopuyo/sudoku_online/backend/hub"
	"github.com/ryo02puyopuyo/sudoku_online/backend/middleware"
	"github.com/ryo02puyopuyo/sudoku_online/backend/room"
)

func main() {
	godotenv.Load()
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}

	roomManager := room.NewRoomManager(5)
	hubInstance := hub.NewHub(roomManager, dbConn)
	apiHandlers := &api.API{DB: dbConn, RoomManager: roomManager}
	authMiddleware := &middleware.Auth{DB: dbConn}

	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/register", apiHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", apiHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/test", apiHandlers.TestHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms", apiHandlers.RoomListHandler).Methods("GET", "OPTIONS")
	apiRouter.Handle("/me", authMiddleware.Optional(http.HandlerFunc(apiHandlers.MeHandler))).Methods("GET", "OPTIONS")

	r.Handle("/ws", authMiddleware.Optional(http.HandlerFunc(hubInstance.ServeWs)))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOrigin := os.Getenv("CORS_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(r)

	go startAdminCLI(hubInstance)

	log.Printf("サーバーがポート %s で起動しました (Origin: %s)", port, allowedOrigin)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startAdminCLI(h *hub.Hub) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Admin CLI 起動完了。コマンドを入力してください (help で一覧表示)")

	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "help":
			fmt.Println("利用可能なコマンド: status, players, msg <text>, reset")

		case "status":
			fmt.Printf("[DEBUG] 接続数: %d\n", h.GetConnectionCount())

		case "players":
			players := h.GetPlayerList()
			fmt.Println("--- Connected Players ---")
			for _, p := range players {
				fmt.Printf("ID: %s | Name: %s | Role: %s | Team: %d\n", p.ID, p.Name, p.Role, p.Team)
			}

		default:
			fmt.Println("不明なコマンドです。")
		}
	}
}
