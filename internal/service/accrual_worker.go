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
	"strconv"
	"sync/atomic"
	"time"
)

const workerCount = 5

var sleepUntil atomic.Int64

func AccrualWorker(ctx context.Context, accURL string, jobs <-chan domain.Order) <-chan domain.Order {
	results := make(chan domain.Order, 1024)

	go func() {
		for i := 0; i < workerCount; i++ {
			worker(ctx, accURL, jobs, results)
		}
	}()

	go func() {
		<-ctx.Done()
		close(results)
	}()

	return results
}

func worker(ctx context.Context, accURL string, jobs <-chan domain.Order, results chan<- domain.Order) {
	for {
		now := time.Now().UnixNano()
		until := sleepUntil.Load()
		if until > now {
			sleepDur := time.Duration(until-now) * time.Nanosecond
			timer := time.NewTimer(sleepDur)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				logger.Log.Warn("resuming work after backoff")
			}
		}

		select {
		case <-ctx.Done():
			return
		case order := <-jobs:
			accRes, retryAfter, err := sendRequest(accURL, order.Number)
			if err != nil {
				logger.Log.Error("error while sending request to accrual system", logger.Error(err))
				continue
			}

			if retryAfter > 0 {
				logger.Log.Warn("accrual system rate limit exceeded, backing off", logger.Int64("seconds", int64(retryAfter)))
				sleepUntil.Store(time.Now().Add(time.Duration(retryAfter) * time.Second).UnixNano())
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

func sendRequest(accURL, number string) (*dto.AccrualResponse, int, error) {
	baseURL, err := url.Parse(accURL)
	if err != nil {
		return nil, 0, err
	}

	baseURL = baseURL.JoinPath("api/orders", number)

	response, err := http.Get(baseURL.String())
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, 0, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			logger.Log.Error("error while closing response body", logger.Error(err))
			return
		}
	}(response.Body)

	if response.StatusCode == http.StatusTooManyRequests {
		retryAfterStr := response.Header.Get("Retry-After")
		retryAfter, _ := strconv.Atoi(retryAfterStr)
		return nil, retryAfter, nil
	}

	var accRes dto.AccrualResponse
	err = json.NewDecoder(response.Body).Decode(&accRes)
	if err != nil {
		return nil, 0, err
	}

	return &accRes, 0, nil
}
