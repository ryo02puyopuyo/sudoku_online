// ゲームで使用するモデル群
package models

type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ClientMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Player struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Team               int    `json:"team"`
	Role               string `json:"role"`
	ConsecutiveCorrect int    `json:"consecutiveCorrect"`
}

type Cell struct {
	Value        int    `json:"value"`
	Status       string `json:"status"`
	FilledByTeam int    `json:"filledByTeam"`
	IsHotSpot    bool   `json:"isHotSpot"`
}

type Score struct {
	Team1 int `json:"team1"`
	Team2 int `json:"team2"`
}

type UserListUpdatePayload struct {
	Players []Player `json:"players"`
	Scores  Score    `json:"scores"`
}

type RoomInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PlayerCount int    `json:"playerCount"`
	Scores      Score  `json:"scores"`
}

type RoomListPayload struct {
	Rooms []RoomInfo `json:"rooms"`
}

type WelcomePayload struct {
	YourPlayer Player     `json:"yourPlayer"`
	BoardState [9][9]Cell `json:"boardState"`
	RoomID     string     `json:"roomId"`
	RoomName   string     `json:"roomName"`
}

type ChatMessage struct {
	SenderName string `json:"senderName"`
	SenderTeam int    `json:"senderTeam"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
}

type GameOverPayload struct {
	WinnerTeam  int   `json:"winnerTeam"`
	FinalScores Score `json:"finalScores"`
}
