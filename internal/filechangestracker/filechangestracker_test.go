package filechangestracker

import (
	"context"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/danielboakye/filechangestracker/internal/config"
	"github.com/danielboakye/filechangestracker/internal/mongolog"
	mongologmock "github.com/danielboakye/filechangestracker/mocks/mongolog"
	osquerymanagermock "github.com/danielboakye/filechangestracker/mocks/osquerymanager"
	"github.com/danielboakye/filechangestracker/pkg/osquerymanager"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v -cover ./pkg/filechangestracker/...

// go test -v -cover -run TestCheckFileChanges ./pkg/filechangestracker
func TestCheckFileChanges(t *testing.T) {
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	mockOSQueryManager := osquerymanagermock.NewMockOSQueryManager(mockCtrl)
	mockMongolog := mongologmock.NewMockLogStore(mockCtrl)

	ctx := context.Background()

	cfg := &config.Config{}
	appLogger := slog.Default()

	tracker := New(appLogger, cfg, mockOSQueryManager, mockMongolog)
	it := tracker.(*fileChangesTracker)

	mockMongolog.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	mockOSQueryManager.EXPECT().Query(gomock.Any()).Return([]map[string]string{
		{
			"target_path": "test/test.txt",
			"time":        timeStr,
		},
	}, nil).AnyTimes()

	mockMongolog.EXPECT().ReadLogsPaginated(gomock.Any(), gomock.Any(), gomock.Any()).Return([]mongolog.LogEntry{
		{
			ID: uuid.NewString(),
			Details: map[string]string{
				"target_path": "test/test.txt",
				"time":        timeStr,
			},
		},
	}, nil).Times(1)

	err := it.checkFileChanges(ctx)
	assert.Nil(err)

	res, err := tracker.GetLogs(ctx, 1, 0)
	assert.Nil(err)
	assert.NotNil(res)
	assert.Len(res, 1)
}

// go test -v -cover -run TestHealthCheck ./pkg/filechangestracker
func TestHealthCheck(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mockCtrl := gomock.NewController(t)
	mockOSQueryManager := osquerymanagermock.NewMockOSQueryManager(mockCtrl)
	mockMongolog := mongologmock.NewMockLogStore(mockCtrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{}
	appLogger := slog.Default()

	tracker := New(appLogger, cfg, mockOSQueryManager, mockMongolog)

	mockOSQueryManager.EXPECT().Query(gomock.Any()).Return(nil, osquerymanager.ErrNoChangesFound).AnyTimes()

	err := tracker.Start(ctx)
	require.NoError(err)

	time.Sleep(2 * time.Second)

	isAlive := tracker.IsTimerThreadAlive()
	assert.True(isAlive)
}
