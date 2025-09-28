package listener

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/quic-go/quic-go/http3"
)

type Http3 struct {
	server  *http3.Server
	handler http.Handler

	waitRequest sync.WaitGroup
	activeCount atomic.Int32

	certFile string
	keyFile  string
	caFile   string
}

// ActiveRequests implements AppListener.
func (s *Http3) ActiveRequests() int {
	return int(s.activeCount.Load())
}

// ListenAndServe implements AppListener.
func (s *Http3) ListenAndServe() error {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.waitRequest.Add(1)
		s.activeCount.Add(1)
		defer func() {
			s.activeCount.Add(-1)
			s.waitRequest.Done()
		}()

		s.handler.ServeHTTP(w, r)
	})

	tlsConfig, err := createTLSConfig(s.certFile, s.keyFile, s.caFile)
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	s.server.Handler = wrappedHandler
	s.server.TLSConfig = tlsConfig

	// Start serving
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown implements AppListener.
func (s *Http3) Shutdown(timeout time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("[HTTP3] Initiating graceful shutdown for app at %s\n", s.server.Addr)
	shutdownErr := s.server.Shutdown(ctx)

	done := make(chan struct{})
	go func() {
		s.waitRequest.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		if shutdownErr == nil {
			return ctx.Err()
		}
	}

	return shutdownErr
}

var _ AppListener = (*Http3)(nil)

func NewHttp3(config map[string]any, handler http.Handler) AppListener {
	idleTimeout := utils.GetValueFromMap(config, IDLE_TIMEOUT_KEY, DEFAULT_IDLE_TIMEOUT)
	addr := utils.GetValueFromMap(config, "addr", ":8080")

	certFile := utils.GetValueFromMap(config, CERT_FILE_KEY, "")
	keyFile := utils.GetValueFromMap(config, KEY_FILE_KEY, "")
	caFile := utils.GetValueFromMap(config, CA_FILE_KEY, "")

	return &Http3{
		handler:  handler,
		certFile: certFile,
		keyFile:  keyFile,
		caFile:   caFile,
		server: &http3.Server{
			Addr:        addr,
			IdleTimeout: idleTimeout,
		},
	}
}
