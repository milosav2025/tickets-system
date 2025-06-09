package services

import (
	"fmt"
	"goticketsistem/db"
	"goticketsistem/models"
	"goticketsistem/utils"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type SystemTicketService struct {
	db *db.DBManager
}

func NewSystemTicketService(db *db.DBManager) *SystemTicketService {
	return &SystemTicketService{db: db}
}

func (sts *SystemTicketService) ProcessSystemTicket(ticketID int, ticket *models.Ticket) error {
	tx, err := sts.db.BeginTransaction()
	if err != nil {
		return err
	}

	var fixedIDs, freeIDs []int
	var oddsMap = make(map[int]float64)
	rows, err := tx.Query(`SELECT selection_id, odd_value, is_fixed FROM selections WHERE ticket_id = $1`, ticketID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var odd float64
		var isFixed bool
		if err := rows.Scan(&id, &odd, &isFixed); err != nil {
			tx.Rollback()
			return err
		}
		oddsMap[id] = odd
		if isFixed {
			fixedIDs = append(fixedIDs, id)
		} else {
			freeIDs = append(freeIDs, id)
		}
	}
	log.Printf("Fixed IDs: %v (count: %d), Free IDs: %v (count: %d), OddsMap: %v", fixedIDs, len(fixedIDs), freeIDs, len(freeIDs), oddsMap)

	systemCombos := strings.Split(strings.TrimSpace(ticket.SystemCombination), ",")
	numCombinations := 0
	for _, combo := range systemCombos {
		parts := strings.Split(strings.TrimSpace(combo), "/")
		k, err := parseInt(parts[0]) // Broj slobodnih selekcija za izbor
		if err != nil {
			tx.Rollback()
			return err
		}
		// Generiši kombinacije samo iz slobodnih selekcija
		binomResult := utils.Binom(len(freeIDs), k)
		numCombinations += binomResult
		log.Printf("Combo: %s, k: %d, len(freeIDs): %d, Binom(%d, %d) = %d", combo, k, len(freeIDs), len(freeIDs), k, binomResult)
	}
	log.Printf("Calculated numCombinations: %d", numCombinations)

	if numCombinations == 0 {
		tx.Rollback()
		return fmt.Errorf("no valid combinations calculated")
	}

	stakePerCombination := ticket.TotalStake / float64(numCombinations)
	log.Printf("Stake per combination: %f", stakePerCombination)
	var maxPayout, minPayout float64 = 0, math.MaxFloat64

	// Generiši sve kombinacije
	for _, combo := range systemCombos {
		parts := strings.Split(strings.TrimSpace(combo), "/")
		k, err := parseInt(parts[0])
		if err != nil {
			tx.Rollback()
			return err
		}
		log.Printf("Processing combo %s, k=%d", combo, k)
		freeCombinations := utils.GenerateCombinations(freeIDs, k)
		log.Printf("Generated %d free combinations for k=%d (expected: %d)", len(freeCombinations), k, utils.Binom(len(freeIDs), k))
		if len(freeCombinations) == 0 {
			log.Printf("Warning: No combinations generated for combo %s", combo)
		}
		for i, comboIDs := range freeCombinations {
			finalCombo := append(fixedIDs, comboIDs...)
			odds := calculateOdds(finalCombo, oddsMap)
			potentialWin := odds * stakePerCombination
			log.Printf("Combo %d: IDs=%v, Odds=%f, PotentialWin=%f", i, finalCombo, odds, potentialWin)

			// Akumulacija max_payout
			maxPayout += potentialWin
			// Akumulacija min_payout
			if potentialWin < minPayout {
				minPayout = potentialWin
			}

			stmt := `INSERT INTO combinations (ticket_id, selection_ids, combination_odds, stake_per_combination, potential_win, status, created_at)
                     VALUES ($1, $2, $3, $4, $5, $6, $7)`
			if _, err := tx.Exec(stmt, ticketID, pq.Array(finalCombo), odds, stakePerCombination, potentialWin, "pending", time.Now()); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	log.Printf("Final maxPayout: %f, minPayout: %f", maxPayout, minPayout)

	// Provera i ažuriranje baze
	updateStmt := `UPDATE tickets SET num_combinations = $1, max_payout = $2, min_payout = $3 WHERE ticket_id = $4`
	result, err := tx.Exec(updateStmt, numCombinations, maxPayout, minPayout, ticketID)
	if err != nil {
		tx.Rollback()
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	log.Printf("Rows affected by UPDATE: %d", rowsAffected)
	if rowsAffected == 0 {
		log.Printf("Warning: No rows updated for ticket_id %d", ticketID)
	} else {
		log.Printf("Updated max_payout: %f, min_payout: %f for ticket_id %d", maxPayout, minPayout, ticketID)
	}

	return tx.Commit()
}

// Pomoćne funkcije
func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func findMinK(combinations []string) int {
	minK := int(^uint(0) >> 1) // MaxInt
	for _, combo := range combinations {
		parts := strings.Split(strings.TrimSpace(combo), "/")
		k, _ := parseInt(parts[0])
		if k < minK {
			minK = k
		}
	}
	return minK
}

func calculateOdds(comboIDs []int, oddsMap map[int]float64) float64 {
	odds := 1.0
	for _, id := range comboIDs {
		odds *= oddsMap[id]
	}
	return odds
}
