package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
	"ticket-reservation/internal/service"
)

type WaitingListHandler struct {
    waitingListService *service.WaitingListService
}

func NewWaitingListHandler(waitingListService *service.WaitingListService) *WaitingListHandler {
    return &WaitingListHandler{waitingListService: waitingListService}
}

func (h *WaitingListHandler) Join(w http.ResponseWriter, r *http.Request) {
    var req struct {
        TicketID string `json:"ticket_id"`
        UserID   string `json:"user_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err := h.waitingListService.AddToWaitingList(r.Context(), req.TicketID, req.UserID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *WaitingListHandler) Leave(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    ticketID := vars["ticketId"]
    userID := vars["userId"]

    err := h.waitingListService.RemoveFromWaitingList(r.Context(), ticketID, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}