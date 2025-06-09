package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"goticketsistem/db"
	"goticketsistem/models"
	"goticketsistem/services"
)

type TicketHandler struct {
	dbManager *db.DBManager
	service   *services.TicketService
}

func NewTicketHandler(dbManager *db.DBManager) *TicketHandler {
	return &TicketHandler{dbManager: dbManager, service: services.NewTicketService(dbManager)}
}

func (th *TicketHandler) HandleTicket(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request on /ticket")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ticket models.Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Decoded ticket: %+v", ticket) // Debug log

	// Privremeno zaobilazimo validaciju user_id
	// if ticket.UserID <= 0 {
	//     http.Error(w, "Invalid user ID", http.StatusBadRequest)
	//     return
	// }
	if ticket.TotalStake <= 0 {
		http.Error(w, "Invalid total stake", http.StatusBadRequest)
		return
	}
	if len(ticket.Selections) == 0 {
		http.Error(w, "No selections provided", http.StatusBadRequest)
		return
	}

	for _, sel := range ticket.Selections {
		if sel.OddValue <= 0 || sel.Stake <= 0 || sel.EventDate.IsZero() {
			http.Error(w, "Invalid selection data", http.StatusBadRequest)
			return
		}
	}

	ticketID, err := th.service.ProcessTicket(&ticket)
	if err != nil {
		log.Printf("Error processing ticket: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]int{"ticket_id": ticketID}); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
