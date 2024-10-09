package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/danielboakye/filechangestracker/internal/mongolog"
	commandexecutormock "github.com/danielboakye/filechangestracker/mocks/commandexecutor"
	filechangestrackermock "github.com/danielboakye/filechangestracker/mocks/filechangestracker"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v -cover ./pkg/httpserver/...

// go test -v -cover -run TestHealthCheck ./pkg/httpserver
func TestHealthCheck(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mockCtrl := gomock.NewController(t)
	mockCmdExecutor := commandexecutormock.NewMockCommandExecutor(mockCtrl)
	mockFileTracker := filechangestrackermock.NewMockFileChangesTracker(mockCtrl)

	appLogger := slog.Default()
	handler := NewHandler(mockFileTracker, mockCmdExecutor)
	router := handler.RegisterRoutes()
	apiServer := NewServer(":9000", appLogger, router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	r.Header.Set("Content-Type", "application/json")

	mockCmdExecutor.EXPECT().IsWorkerThreadAlive().Return(true).Times(1)
	mockFileTracker.EXPECT().IsTimerThreadAlive().Return(true).Times(1)

	apiServer.httpServer.Handler.ServeHTTP(w, r)

	assert.Equal(http.StatusOK, w.Code)

	res := HealthCheckResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(err)

	assert.True(res.TimerThread)
	assert.True(res.WorkerThread)
}

// go test -v -cover -run TestSubmitCommands ./pkg/httpserver
func TestSubmitCommands(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mockCtrl := gomock.NewController(t)
	mockCmdExecutor := commandexecutormock.NewMockCommandExecutor(mockCtrl)
	mockFileTracker := filechangestrackermock.NewMockFileChangesTracker(mockCtrl)

	appLogger := slog.Default()
	handler := NewHandler(mockFileTracker, mockCmdExecutor)
	router := handler.RegisterRoutes()
	apiServer := NewServer(":9000", appLogger, router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/commands", strings.NewReader(`{"commands":["touch /Users/user/Downloads/test/test.txt"]}`))
	r.Header.Set("Content-Type", "application/json")

	mockCmdExecutor.EXPECT().AddCommands(gomock.Any()).Return(nil).Times(1)

	apiServer.httpServer.Handler.ServeHTTP(w, r)

	assert.Equal(http.StatusOK, w.Code)

	res := map[string]string{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(err)

	msg, exists := res["message"]
	assert.True(exists)
	assert.Equal(msg, "commands added to queue")
}

// go test -v -cover -run TestSubmitCommands_Failed ./pkg/httpserver
func TestSubmitCommands_Failed(t *testing.T) {
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	mockCmdExecutor := commandexecutormock.NewMockCommandExecutor(mockCtrl)
	mockFileTracker := filechangestrackermock.NewMockFileChangesTracker(mockCtrl)

	appLogger := slog.Default()
	handler := NewHandler(mockFileTracker, mockCmdExecutor)
	router := handler.RegisterRoutes()
	apiServer := NewServer(":9000", appLogger, router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/commands", strings.NewReader(`{"commands": "touch /Users/user/Downloads/test/test.txt"}`))
	r.Header.Set("Content-Type", "application/json")

	apiServer.httpServer.Handler.ServeHTTP(w, r)

	assert.Equal(http.StatusBadRequest, w.Code)
}

// go test -v -cover -run TestGetLogs ./pkg/httpserver
func TestGetLogs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mockCtrl := gomock.NewController(t)
	mockCmdExecutor := commandexecutormock.NewMockCommandExecutor(mockCtrl)
	mockFileTracker := filechangestrackermock.NewMockFileChangesTracker(mockCtrl)

	appLogger := slog.Default()
	handler := NewHandler(mockFileTracker, mockCmdExecutor)
	router := handler.RegisterRoutes()
	apiServer := NewServer(":9000", appLogger, router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/logs", nil)
	r.Header.Set("Content-Type", "application/json")

	mockFileTracker.EXPECT().GetLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return([]mongolog.LogEntry{
		{
			ID: uuid.NewString(),
			Details: map[string]string{
				"target_path": "test/test.txt",
				"time":        strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
	}, nil).Times(1)

	apiServer.httpServer.Handler.ServeHTTP(w, r)

	assert.Equal(http.StatusOK, w.Code)

	res := []map[string]interface{}{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(err)
	assert.Len(res, 1)
}

// go test -v -cover -run TestNotFound ./pkg/httpserver
func TestNotFound(t *testing.T) {
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	mockCmdExecutor := commandexecutormock.NewMockCommandExecutor(mockCtrl)
	mockFileTracker := filechangestrackermock.NewMockFileChangesTracker(mockCtrl)

	appLogger := slog.Default()
	handler := NewHandler(mockFileTracker, mockCmdExecutor)
	router := handler.RegisterRoutes()
	apiServer := NewServer(":9000", appLogger, router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/log", nil)
	r.Header.Set("Content-Type", "application/json")

	apiServer.httpServer.Handler.ServeHTTP(w, r)

	assert.Equal(http.StatusNotFound, w.Code)
}
