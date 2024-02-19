package reconciler

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

// APIKeys is a map of API keys
type APIKeys map[string]string

// MarshalBinary marshals APIKeys to JSON encoding
func (ak APIKeys) MarshalBinary() ([]byte, error) {
	return json.Marshal(ak)
}

func storeAPIKeys(ctx context.Context, cfg Config, rc *redis.Client) (APIKeys, error) {
	apiKeys := map[string]string{
		"DIODE_TO_NETBOX":     cfg.DiodeToNetBoxAPIKey,
		"NETBOX_TO_DIODE":     cfg.NetBoxToDiodeAPIKey,
		"DATASOURCE_TO_DIODE": cfg.DatasourceToDiodeAPIKey,
	}

	if err := rc.HSet(ctx, "diode.api_keys", apiKeys).Err(); err != nil {
		return nil, err
	}

	return apiKeys, nil
}
