package services

import (
	"fmt"
	"goticketsistem/db"
	"goticketsistem/models"
	"log"
	"time"

	"github.com/lib/pq"
)

type TicketService struct {
	db *db.DBManager
}

func NewTicketService(db *db.DBManager) *TicketService {
	return &TicketService{db: db}
}

func (ts *TicketService) CreateTicket(ticket *models.Ticket) (int, error) {
	tx, err := ts.db.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %v", err)
	}

	if ticket.Status == "" {
		ticket.Status = "pending"
	}
	if ticket.CreatedAt.IsZero() {
		ticket.CreatedAt = time.Now()
	}
	if ticket.Hits < 0 {
		ticket.Hits = 0
	}
	if ticket.Misses < 0 {
		ticket.Misses = 0
	}
	if ticket.Pending < 0 {
		ticket.Pending = 0
	}

	var ticketID int
	stmt := `INSERT INTO tickets (user_id, total_stake, total_odd, potential_payout, hits, misses, pending, status, 
             created_at, max_payout, min_payout, final_payout, num_combinations, system_combination, ticket_type)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING ticket_id`
	err = tx.QueryRow(stmt, ticket.UserID, ticket.TotalStake, ticket.TotalOdd, ticket.PotentialPayout, ticket.Hits,
		ticket.Misses, ticket.Pending, ticket.Status, ticket.CreatedAt, ticket.MaxPayout, ticket.MinPayout,
		ticket.FinalPayout, ticket.NumCombinations, ticket.SystemCombination, ticket.TicketType).Scan(&ticketID)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to insert ticket: %v", err)
	}

	for _, sel := range ticket.Selections {
		stmt := `INSERT INTO selections (ticket_id, sport_type, league, home_team, away_team, event_date, 
                       market_type, selected_outcome, odd_value, stake, eid, selection_type, status, is_fixed)
                       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
		_, err = tx.Exec(stmt, ticketID, sel.SportType, sel.League, sel.HomeTeam, sel.AwayTeam, sel.EventDate,
			sel.MarketType, sel.SelectedOutcome, sel.OddValue, sel.Stake, sel.Eid, sel.SelectionType, "pending", sel.IsFixed)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert selection: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return ticketID, nil
}

func (ts *TicketService) ProcessTicket(ticket *models.Ticket) (int, error) {
	ticketID, err := ts.CreateTicket(ticket)
	if err != nil {
		return 0, err
	}

	log.Printf("Processing ticket %d, type: %s, system_combination: %s", ticketID, ticket.TicketType, ticket.SystemCombination)
	if ticket.TicketType == "system" && ticket.SystemCombination != "" {
		systemService := NewSystemTicketService(ts.db)
		return ticketID, systemService.ProcessSystemTicket(ticketID, ticket)
	}
	return ticketID, ts.processNormalTicket(ticketID, ticket)
}

func (ts *TicketService) processNormalTicket(ticketID int, ticket *models.Ticket) error {
	tx, err := ts.db.BeginTransaction()
	if err != nil {
		return err
	}

	var selectionIDs []int
	var odds float64 = 1.0
	rows, err := tx.Query(`SELECT selection_id, odd_value FROM selections WHERE ticket_id = $1`, ticketID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var odd float64
		if err := rows.Scan(&id, &odd); err != nil {
			tx.Rollback()
			return err
		}
		selectionIDs = append(selectionIDs, id)
		odds *= odd
	}

	potentialWin := odds * ticket.TotalStake
	maxPayout := potentialWin
	minPayout := 0.0 // Za normalni tiket, min je 0 jer sve mora proći
	numCombinations := 1

	stmt := `INSERT INTO combinations (ticket_id, selection_ids, combination_odds, stake_per_combination, potential_win, status, created_at)
             VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.Exec(stmt, ticketID, pq.Array(selectionIDs), odds, ticket.TotalStake, potentialWin, "pending", time.Now()); err != nil {
		tx.Rollback()
		return err
	}

	updateStmt := `UPDATE tickets SET total_odd = $1, potential_payout = $2, max_payout = $3, min_payout = $4, num_combinations = $5 WHERE ticket_id = $6`
	if _, err := tx.Exec(updateStmt, odds, potentialWin, maxPayout, minPayout, numCombinations, ticketID); err != nil {
		tx.Rollback()
		return err
	}

	log.Printf("Processed normal ticket %d, max_payout: %f, min_payout: %f, num_combinations: %d", ticketID, maxPayout, minPayout, numCombinations)
	return tx.Commit()
}
