package balancehandler

import (
	"encoding/json"
	"errors"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/dto"
	"github.com/koyif/gophermart/pkg/logger"
	"github.com/theplant/luhn"
	"net/http"
	"strconv"
	"time"
)

type balanceService interface {
	Balance(userID int64) (*domain.Balance, error)
	Withdraw(orderNumber string, sum float64, userID int64) error
	Withdrawals(userID int64) ([]domain.Withdrawal, error)
}

type BalanceHandler struct {
	balanceService balanceService
}

func New(svc balanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: svc,
	}
}

func (h BalanceHandler) Balance(w http.ResponseWriter, r *http.Request) {
	userIDHeader := r.Header.Get("User-ID")
	userID, err := strconv.ParseInt(userIDHeader, 10, 64)
	if err != nil {
		logger.Log.Error("error while parsing user ID from header", logger.String("user_id", userIDHeader), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	balance, err := h.balanceService.Balance(userID)
	if err != nil {
		logger.Log.Error("error while fetching balance", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := dto.Balance{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Log.Error("error while encoding balance to JSON", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userIDHeader := r.Header.Get("User-ID")
	userID, err := strconv.ParseInt(userIDHeader, 10, 64)
	if err != nil {
		logger.Log.Error("error while parsing user ID from header", logger.String("user_id", userIDHeader), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var withdrawalRequest dto.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&withdrawalRequest); err != nil {
		logger.Log.Warn("error while decoding a withdrawal request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.ParseInt(withdrawalRequest.Order, 10, 64)
	if err != nil {
		logger.Log.Warn("invalid order ID", logger.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if ok := luhn.Valid(int(orderNumber)); !ok {
		logger.Log.Warn("invalid order ID, Luhn check failed", logger.Int64("orderNumber", orderNumber))
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err = h.balanceService.Withdraw(withdrawalRequest.Order, withdrawalRequest.Sum, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientFunds) {
			logger.Log.Warn("insufficient funds", logger.Int64("user_id", userID))
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
			return
		} else if errors.Is(err, domain.ErrWithdrawalExists) {
			logger.Log.Warn("withdrawal already exists", logger.Int64("user_id", userID))
			http.Error(w, "withdrawal already exists", http.StatusOK)
			return
		} else if errors.Is(err, domain.ErrWithdrawalAddedByAnotherUser) {
			logger.Log.Warn("withdrawal belongs to another user", logger.Int64("user_id", userID))
			http.Error(w, "withdrawal belongs to another user", http.StatusConflict)
			return
		}

		logger.Log.Error("error while withdrawing money", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (h BalanceHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userIDHeader := r.Header.Get("User-ID")
	userID, err := strconv.ParseInt(userIDHeader, 10, 64)
	if err != nil {
		logger.Log.Error("error while parsing user ID from header", logger.String("user_id", userIDHeader), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	withdrawals, err := h.balanceService.Withdrawals(userID)
	if err != nil {
		logger.Log.Error("error while fetching withdrawals", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dtos := make([]dto.Withdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		dtos[i] = dto.Withdrawal{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Amount,
			ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(dtos)
	if err != nil {
		logger.Log.Error("error while encoding withdrawals to JSON", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
