package service

import (
	"context"
	"encoding/json"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/dto"
	"github.com/koyif/gophermart/pkg/logger"
	"io"
	"net/http"
	"net/url"
)

const workerCount = 5

func AccrualWorker(ctx context.Context, accURL string, jobs <-chan domain.Order) <-chan domain.Order {
	results := make(chan domain.Order, 1024)

	go func() {
		for i := 0; i < workerCount; i++ {
			worker(ctx, accURL, jobs, results)
		}
	}()

	return results
}

func worker(ctx context.Context, accURL string, jobs <-chan domain.Order, results chan<- domain.Order) {
	for {
		select {
		case <-ctx.Done():
			return
		case order := <-jobs:
			accRes, err := sendRequest(accURL, order.Number)
			if err != nil {
				logger.Log.Error("error while sending request to accrual system", logger.Error(err))
				continue
			}
			if accRes.Status != "" && accRes.Status != order.Status {
				order.Status = accRes.Status
				order.Accrual = accRes.Accrual
				results <- order
			}
		}
	}
}

func sendRequest(accURL, number string) (*dto.AccrualResponse, error) {
	baseUrl, err := url.Parse(accURL)
	if err != nil {
		return nil, err
	}

	baseUrl = baseUrl.JoinPath("api/orders", number)

	response, err := http.Get(baseUrl.String())
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Log.Error("error while closing response body", logger.Error(err))
			return
		}
	}(response.Body)

	var accRes dto.AccrualResponse
	err = json.NewDecoder(response.Body).Decode(&accRes)
	if err != nil {
		return nil, err
	}

	return &accRes, nil
}
