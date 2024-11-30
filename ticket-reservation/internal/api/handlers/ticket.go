package handlers

import (
    "encoding/json"
    "net/http"
    "ticket-reservation/internal/domain"
    "ticket-reservation/internal/service"
)

type TicketHandler struct {
    ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
    return &TicketHandler{ticketService: ticketService}
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tickets, err := h.ticketService.List(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(tickets)
}

func (h *TicketHandler) Reserve(w http.ResponseWriter, r *http.Request) {
    var req domain.ReservationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ctx := r.Context()
    err := h.ticketService.ReserveTicket(ctx, &req)
    if err != nil {
        status := http.StatusInternalServerError
        if err == service.ErrTicketAlreadyReserved {
            status = http.StatusConflict
        }
        http.Error(w, err.Error(), status)
        return
    }

    w.WriteHeader(http.StatusOK)
}