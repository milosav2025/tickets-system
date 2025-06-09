package models

import "time"

type Ticket struct {
	UserID            int
	TotalStake        float64
	TotalOdd          float64
	PotentialPayout   float64
	Hits              int
	Misses            int
	Pending           int
	Status            string
	CreatedAt         time.Time
	MaxPayout         float64
	MinPayout         float64
	FinalPayout       float64
	NumCombinations   int
	SystemCombination string
	TicketType        string
	Selections        []Selection
	Logo              string
}

type DBTicket struct {
	TicketID          int
	UserID            int
	TotalStake        float64
	TotalOdd          float64
	PotentialPayout   float64
	Hits              int
	Misses            int
	Pending           int
	Status            string
	CreatedAt         time.Time
	MaxPayout         float64
	MinPayout         float64
	FinalPayout       float64
	NumCombinations   int
	SystemCombination *string
	TicketType        string
}

type Selection struct {
	SportType       string
	League          string
	HomeTeam        string
	AwayTeam        string
	EventDate       time.Time
	MarketType      string
	SelectedOutcome string
	OddValue        float64
	Stake           float64
	Eid             string
	SelectionType   string
	Status          string
	IsFixed         bool
}

type DBSelection struct {
	ID              int
	TicketID        int
	SportType       string
	League          string
	HomeTeam        string
	AwayTeam        string
	EventDate       time.Time
	MarketType      string
	SelectedOutcome string
	OddValue        float64
	Stake           float64
	Eid             string
	SelectionType   string
	Status          string
	IsFixed         bool
}

type DBCombination struct {
	CombinationID       int
	TicketID            int
	SelectionIDs        []int
	CombinationOdds     float64
	StakePerCombination float64
	PotentialWin        float64
	Status              string
	FinalPayout         float64
	CreatedAt           time.Time
}
