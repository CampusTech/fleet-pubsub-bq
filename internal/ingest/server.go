package ingest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Server wires the HTTP mux to the ingest handler.
type Server struct {
	handler *Handler
	log     *logrus.Logger
	// subscription ID suffixes for routing
	resultSub string
	statusSub string
	auditSub  string
}

func NewServer(h *Handler, resultSub, statusSub, auditSub string, log *logrus.Logger) *Server {
	return &Server{
		handler:   h,
		log:       log,
		resultSub: resultSub,
		statusSub: statusSub,
		auditSub:  auditSub,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ingest", s.handleIngest)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		s.log.WithError(err).Error("decode pubsub message")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	insertedAt := time.Now().UTC()
	subID := subscriptionSuffix(msg.Subscription)

	log := s.log.WithFields(logrus.Fields{
		"subscription": subID,
		"message_id":   msg.Message.MessageID,
	})

	var err error
	switch {
	case strings.EqualFold(subID, s.resultSub):
		err = s.handler.InsertResult(r.Context(), msg, insertedAt)
	case strings.EqualFold(subID, s.statusSub):
		err = s.handler.InsertStatus(r.Context(), msg, insertedAt)
	case strings.EqualFold(subID, s.auditSub):
		err = s.handler.InsertAudit(r.Context(), msg, insertedAt)
	default:
		log.Warn("unknown subscription; ignoring")
		// Return 200 so PubSub doesn't retry unknown subscriptions.
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		log.WithError(err).Error("ingest failed")
		// 500 causes PubSub to retry with backoff.
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
