package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/prometheus/common/model"
)

var errEndpointRequired = errors.New("the endpoint setting is empty")

const (
	XScopeOrgID = "X-Scope-OrgID"
)

// ClientSettings contain information for the Loki server that the client connects to
type ClientSettings struct {
	// Endpoint of the Loki server.
	URL utils.EnvString `json:"url" yaml:"url"`
	// Headers specify headers to inject in the requests.
	Headers map[string]utils.EnvString `json:"headers" yaml:"headers"`
	// The default timeout in seconds of client requests. The zero value is no timeout.
	Timeout uint `json:"timeout" yaml:"timeout"`
	// The duration to limit the max time range in query requests. If the value is 0 the client will leave the time range validation to Loki server.
	MaxTimeRange model.Duration `json:"max_time_range" yaml:"max_time_range" jsonschema:"type=string,pattern=[0-9]+(h|d|w|y)"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (cs *ClientSettings) UnmarshalJSON(b []byte) error {
	type Plain ClientSettings
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}

	u, err := plain.URL.Get()
	if err != nil || u == "" {
		return fmt.Errorf("invalid client URL %s", err)
	}

	*cs = ClientSettings(plain)

	return nil
}
