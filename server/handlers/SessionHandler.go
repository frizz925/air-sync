package handlers

import (
	"air-sync/models"
	repos "air-sync/repositories"
	"air-sync/util"
	"air-sync/util/logging"
	"air-sync/util/pubsub"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

var (
	ErrSessionNotFound = errors.New("Session not found")
	ResSessionNotFound = &util.RestResponse{
		StatusCode: http.StatusNotFound,
		Message:    "Resource not found",
		Error:      "Session not found",
	}
)

type SessionHandler struct {
	repo   repos.SessionRepository
	stream *pubsub.Stream
}

type SessionHandlerFunc func(req *http.Request, session *models.Session) (interface{}, error)

func NewSessionHandler(repo repos.SessionRepository, stream *pubsub.Stream) *SessionHandler {
	return &SessionHandler{
		repo:   repo,
		stream: stream,
	}
}

func (h *SessionHandler) CreateSessionLogger(req *http.Request, session *models.Session) *log.Logger {
	logger := util.CreateRequestLogger(req)
	h.ApplySessionLogger(logger, session)
	return logger
}

func (h *SessionHandler) ApplySessionLogger(logger *log.Logger, session *models.Session) {
	logger.Formatter = logging.NewSessionLogFormatter(logger.Formatter, session)
}

func (h *SessionHandler) WrapSessionHandlerFunc(handler SessionHandlerFunc) http.HandlerFunc {
	return util.WrapRestHandlerFunc(func(req *http.Request) (*util.RestResponse, error) {
		id := mux.Vars(req)["id"]
		session, err := h.repo.Get(id)
		if err != nil {
			return h.HandleSessionRestError(err)
		}
		h.ApplySessionLogger(util.RequestLogger(req), session)
		data, err := handler(req, session)
		if err != nil {
			return nil, err
		}
		return &util.RestResponse{
			Data: data,
		}, nil
	})
}

func (h *SessionHandler) HandleSessionRestError(err error) (*util.RestResponse, error) {
	if err == repos.ErrSessionNotFound {
		return ResSessionNotFound, nil
	}
	return nil, err
}

func (h *SessionHandler) HandleSessionError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if err == repos.ErrSessionNotFound {
		code = http.StatusNotFound
	}
	http.Error(w, err.Error(), code)
}
