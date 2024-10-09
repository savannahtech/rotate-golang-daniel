package osquerymanager

import (
	"fmt"

	"github.com/osquery/osquery-go"
)

var ErrNoChangesFound = fmt.Errorf("no matches found")

//go:generate mockgen -destination=../../mocks/osquerymanager/mock_osquerymanager.go -package=osquerymanagermock -source=osquerymanager.go
type OSQueryManager interface {
	Query(sql string) ([]map[string]string, error)
	Close() error
}

type osQueryManager struct {
	osqueryClient *osquery.ExtensionManagerClient
}

// osquery.ExtensionPluginResponse
func New(osqueryClient *osquery.ExtensionManagerClient) OSQueryManager {
	return &osQueryManager{
		osqueryClient: osqueryClient,
	}
}

func (m *osQueryManager) Query(sql string) ([]map[string]string, error) {
	res, err := m.osqueryClient.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("error running osquery: %w", err)
	}
	if res.Status.Code != 0 {
		return nil, fmt.Errorf("error running osquery: %s", res.Status.Message)
	}
	if len(res.Response) == 0 {
		return nil, ErrNoChangesFound
	}

	return res.Response, nil
}

func (m *osQueryManager) Close() error {
	if m.osqueryClient != nil {
		m.osqueryClient.Close()
		m.osqueryClient = nil
	}

	return nil
}
