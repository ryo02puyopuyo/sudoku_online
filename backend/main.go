package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	hubInstance := hub.NewHub(gameInstance, dbConn) // 修正点: dbConnを渡す
	apiHandlers := &api.API{DB: dbConn}
	authMiddleware := &middleware.Auth{DB: dbConn}

	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/register", apiHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", apiHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/test", apiHandlers.TestHandler).Methods("POST", "OPTIONS")
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

	//デバッグツール
	go startAdminCLI(hubInstance, gameInstance) // goroutineで起動

	//start server
	log.Println("サーバーがポート8080で起動しました")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startAdminCLI(h *hub.Hub, g *game.Game) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Admin CLI 起動完了。コマンドを入力してください (help で一覧表示)")

	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		//args := parts[1:]

		switch command {
		case "help":
			fmt.Println("利用可能なコマンド: status, players, msg <text>, reset")

		case "status":
			// ゲーム全体の状況を表示
			scores := g.GetScores()
			fmt.Printf("[DEBUG] 接続数: %d | スコア: T1=%d, T2=%d\n", h.GetConnectionCount(), scores.Team1, scores.Team2)

		case "players":
			// 接続中のプレイヤー詳細をターミナルに一覧表示
			// ※後述する Hub へのメソッド追加が必要
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
