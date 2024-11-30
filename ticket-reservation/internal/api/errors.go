package api

import (
    "encoding/json"
    "errors"
    "net/http"
	"ticket-reservation/internal/domain"
	"ticket-reservation/internal/service"
)

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Message string `json:"message"`
}

var (
    ErrTicketNotFound        = errors.New("ticket not found")
    ErrTicketAlreadyReserved = errors.New("ticket already reserved")
    ErrInvalidRequest        = errors.New("invalid request")
    ErrReservationExpired    = errors.New("reservation expired")
    ErrUserNotInWaitingList  = errors.New("user not in waiting list")
)

func WriteError(w http.ResponseWriter, err error, status int) {
    response := ErrorResponse{
        Error: err.Error(),
        Code:  getErrorCode(err),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}

func getErrorCode(err error) string {
    switch {
    case errors.Is(err, ErrTicketNotFound):
        return "TICKET_NOT_FOUND"
    case errors.Is(err, ErrTicketAlreadyReserved):
        return "TICKET_ALREADY_RESERVED"
    case errors.Is(err, ErrInvalidRequest):
        return "INVALID_REQUEST"
    case errors.Is(err, ErrReservationExpired):
        return "RESERVATION_EXPIRED"
    case errors.Is(err, ErrUserNotInWaitingList):
        return "USER_NOT_IN_WAITING_LIST"
    default:
        return "INTERNAL_SERVER_ERROR"
    }
}

// Example usage in handler:
func (h *TicketHandler) Reserve(w http.ResponseWriter, r *http.Request) {
    var req ReservationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
        return
    }

    err := h.ticketService.ReserveTicket(r.Context(), &domain.ReservationRequest{
        TicketID: req.TicketID,
        UserID:   req.UserID,
    })
    if err != nil {
        switch {
        case errors.Is(err, ErrTicketNotFound):
            WriteError(w, err, http.StatusNotFound)
        case errors.Is(err, ErrTicketAlreadyReserved):
            WriteError(w, err, http.StatusConflict)
        default:
            WriteError(w, err, http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
}

type TicketHandler struct {
    ticketService service.TicketService
}

type ReservationRequest struct {
    TicketID string `json:"ticket_id"`
    UserID   string `json:"user_id"`
}