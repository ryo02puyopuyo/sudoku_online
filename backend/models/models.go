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
	ID   string `json:"id"`
	Name string `json:"name"`
	Team int    `json:"team"`
}
type Cell struct {
	Value        int    `json:"value"`
	Status       string `json:"status"`
	FilledByTeam int    `json:"filledByTeam"`
}
type Score struct {
	Team1 int `json:"team1"`
	Team2 int `json:"team2"`
}
type UserListUpdatePayload struct {
	Players []Player `json:"players"`
	Scores  Score    `json:"scores"`
}

type WelcomePayload struct {
	YourPlayer Player     `json:"yourPlayer"`
	BoardState [9][9]Cell `json:"boardState"`
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
