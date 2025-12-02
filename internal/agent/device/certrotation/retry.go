package certrotation

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/flightctl/flightctl/internal/agent/config"
	"github.com/sirupsen/logrus"
)

// RetryQueue manages retry scheduling for failed renewal requests
type RetryQueue struct {
	config    *config.CertRotationConfig
	requests  map[string]*RenewalRequest // requestID -> RenewalRequest
	mu        sync.RWMutex
	ticker    *time.Ticker
	stopCh    chan struct{}
	renewalFn func(context.Context, *RenewalRequest) error
	log       *logrus.Logger
}

// NewRetryQueue creates a new retry queue for certificate renewal requests
func NewRetryQueue(cfg *config.CertRotationConfig, renewalFn func(context.Context, *RenewalRequest) error, log *logrus.Logger) *RetryQueue {
	return &RetryQueue{
		config:    cfg,
		requests:  make(map[string]*RenewalRequest),
		stopCh:    make(chan struct{}),
		renewalFn: renewalFn,
		log:       log,
	}
}

// Start begins processing the retry queue
func (rq *RetryQueue) Start(ctx context.Context) {
	// Check retry queue every minute
	rq.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				rq.log.Info("Retry queue context cancelled, shutting down")
				return
			case <-rq.stopCh:
				rq.log.Info("Retry queue stopped")
				return
			case <-rq.ticker.C:
				rq.processRetries(ctx)
			}
		}
	}()

	rq.log.Info("Certificate renewal retry queue started")
}

// Stop stops the retry queue
func (rq *RetryQueue) Stop() {
	if rq.ticker != nil {
		rq.ticker.Stop()
	}
	close(rq.stopCh)
}

// Add adds a renewal request to the retry queue with exponential backoff
func (rq *RetryQueue) Add(req *RenewalRequest) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	// Calculate next retry time using exponential backoff
	backoffDelay := rq.calculateBackoff(req.RetryCount)
	req.NextRetryAt = time.Now().Add(backoffDelay)
	req.RetryCount++

	rq.requests[req.RequestID] = req

	rq.log.WithFields(logrus.Fields{
		"request_id":    req.RequestID,
		"retry_count":   req.RetryCount,
		"next_retry_at": req.NextRetryAt,
		"backoff_delay": backoffDelay,
	}).Info("Added renewal request to retry queue")
}

// Remove removes a renewal request from the retry queue
func (rq *RetryQueue) Remove(requestID string) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	delete(rq.requests, requestID)

	rq.log.WithFields(logrus.Fields{
		"request_id": requestID,
	}).Debug("Removed renewal request from retry queue")
}

// processRetries checks for requests that are ready to retry and processes them
func (rq *RetryQueue) processRetries(ctx context.Context) {
	rq.mu.Lock()
	now := time.Now()
	var toRetry []*RenewalRequest

	// Find requests ready to retry
	for _, req := range rq.requests {
		if now.After(req.NextRetryAt) || now.Equal(req.NextRetryAt) {
			toRetry = append(toRetry, req)
		}
	}
	rq.mu.Unlock()

	// Process retries outside the lock to avoid blocking
	for _, req := range toRetry {
		rq.log.WithFields(logrus.Fields{
			"request_id":  req.RequestID,
			"retry_count": req.RetryCount,
			"device_id":   req.DeviceID,
		}).Info("Retrying certificate renewal request")

		// Update status to submitted
		req.Status = StatusSubmitted

		// Attempt renewal
		err := rq.renewalFn(ctx, req)
		if err != nil {
			rq.log.WithFields(logrus.Fields{
				"request_id":  req.RequestID,
				"retry_count": req.RetryCount,
				"error":       err,
			}).Warn("Certificate renewal retry failed, will retry again")

			// Keep in queue with incremented retry count
			req.Status = StatusFailed
			rq.Add(req)
		} else {
			// Success - remove from retry queue
			rq.log.WithFields(logrus.Fields{
				"request_id":  req.RequestID,
				"retry_count": req.RetryCount,
			}).Info("Certificate renewal retry succeeded")

			req.Status = StatusCompleted
			rq.Remove(req.RequestID)
		}
	}
}

// calculateBackoff calculates the exponential backoff delay for a given retry count
func (rq *RetryQueue) calculateBackoff(retryCount int) time.Duration {
	// Exponential backoff: initialInterval * (multiplier ^ retryCount)
	initialInterval := time.Duration(rq.config.RetryInitialInterval)
	maxInterval := time.Duration(rq.config.RetryMaxInterval)
	multiplier := rq.config.RetryBackoffMultiplier

	// Calculate delay with exponential backoff
	delay := float64(initialInterval) * math.Pow(multiplier, float64(retryCount))

	// Cap at maximum interval
	if delay > float64(maxInterval) {
		return maxInterval
	}

	return time.Duration(delay)
}

// Size returns the current number of requests in the retry queue
func (rq *RetryQueue) Size() int {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return len(rq.requests)
}

// Get retrieves a renewal request from the retry queue
func (rq *RetryQueue) Get(requestID string) (*RenewalRequest, bool) {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	req, exists := rq.requests[requestID]
	return req, exists
}
