package orderhandler

import (
	"encoding/json"
	"errors"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/dto"
	"github.com/koyif/gophermart/pkg/logger"
	"github.com/theplant/luhn"
	"io"
	"net/http"
	"strconv"
	"time"
)

type OrderService interface {
	Create(orderID string, userID int64) error
	Orders(userID int64) ([]domain.Order, error)
}

type OrderHandler struct {
	srv OrderService
}

func New(srv OrderService) *OrderHandler {
	return &OrderHandler{
		srv: srv,
	}
}

func (h OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Warn("error while reading request body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			logger.Log.Error("error while closing request body", logger.Error(err))
			return
		}
	}(r.Body)

	if len(body) == 0 {
		logger.Log.Warn("empty request body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid order ID", logger.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if ok := luhn.Valid(int(orderID)); !ok {
		logger.Log.Warn("invalid order ID, Luhn check failed", logger.Int64("order_id", orderID))
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	userIDHeader := r.Header.Get("User-ID")
	userID, err := strconv.ParseInt(userIDHeader, 10, 64)
	if err != nil {
		logger.Log.Error("error while parsing user ID from header", logger.String("user_id", userIDHeader), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.srv.Create(strconv.FormatInt(orderID, 10), userID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderExists) {
			logger.Log.Warn("order already exists", logger.Int64("order_id", orderID))
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, domain.ErrOrderAddedByAnotherUser) {
			logger.Log.Warn("order belongs to another user", logger.Int64("order_id", orderID))
			http.Error(w, "order belongs to another user", http.StatusConflict)
			return
		}
		logger.Log.Error("error while creating order", logger.Int64("order_id", orderID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userIDHeader := r.Header.Get("User-ID")
	userID, err := strconv.ParseInt(userIDHeader, 10, 64)
	if err != nil {
		logger.Log.Error("error while parsing user ID from header", logger.String("user_id", userIDHeader), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	orders, err := h.srv.Orders(userID)
	if err != nil {
		logger.Log.Error("error while fetching orders", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	dtos := make([]dto.Order, len(orders))
	for i, order := range orders {
		dtos[i] = dto.Order{
			Number:     order.OrderID,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(dtos)
	if err != nil {
		logger.Log.Error("error while encoding orders to JSON", logger.Int64("user_id", userID), logger.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
