package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/danielboakye/filechangestracker/pkg/response"
)

// CommandRequest represents the structure of a command request
type CommandRequest struct {
	Commands []string `json:"commands"`
}

// HealthCheckResponse represents the structure of the health check response
type HealthCheckResponse struct {
	WorkerThread bool `json:"worker_thread_alive"`
	TimerThread  bool `json:"timer_thread_alive"`
}

// LogsResponse represents the structure of logs response
type LogsResponse struct {
	Logs []string `json:"logs"`
}

func (h *Handler) HandleSubmitCommands(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.InvalidRequest(w, err.Error())
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		response.InvalidRequest(w, err.Error())
		return
	}
	if len(req.Commands) == 0 {
		response.InvalidRequest(w, "no commands submitted")
		return
	}

	err = h.executor.AddCommands(req.Commands)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "commands added to queue",
	})
}

// handleHealthCheck returns the health status of the worker and timer threads
func (h *Handler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	res := HealthCheckResponse{
		WorkerThread: h.executor.IsWorkerThreadAlive(),
		TimerThread:  h.tracker.IsTimerThreadAlive(),
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *Handler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var offset, limit int64 = 0, 10
	var err error
	queryOffset := q.Get("offset")
	if queryOffset != "" {
		offset, err = strconv.ParseInt(queryOffset, 10, 64)
		if err != nil {
			response.InvalidRequest(w, "offset field is not an integer")
			return
		}
	}

	queryLimit := q.Get("limit")
	if queryLimit != "" {
		limit, err = strconv.ParseInt(queryLimit, 10, 64)
		if err != nil {
			response.InvalidRequest(w, "limit field is not an integer")
			return
		}
	}
	if limit < 1 {
		response.InvalidRequest(w, "limit field cannot be less than 1")
		return
	}

	res, err := h.tracker.GetLogs(r.Context(), limit, offset)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *Handler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusNotFound, map[string]string{
		"message": fmt.Sprintf("resource: (%s) could not be found", r.URL.Path),
	})
}
