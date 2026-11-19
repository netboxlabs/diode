package changeset_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

func TestNewErrorKeepsDetailsAsJSON(t *testing.T) {
	tests := []struct {
		name        string
		details     []byte
		wantDetails string
	}{
		{name: "json object", details: []byte(`{"errors":["boom"]}`), wantDetails: `{"errors":["boom"]}`},
		{name: "json array", details: []byte(`["boom"]`), wantDetails: `["boom"]`},
		{name: "plain text", details: []byte("Selected branch is not ready."), wantDetails: `"Selected branch is not ready."`},
		{name: "html", details: []byte("<html><body>502 Bad Gateway</body></html>"), wantDetails: `"<html><body>502 Bad Gateway</body></html>"`},
		{name: "empty", details: nil, wantDetails: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := changeset.NewError("bulk plan failed", errors.ErrCodeOpsGenerateDiff, tt.details)

			var csErr *changeset.Error
			require.ErrorAs(t, err, &csErr)
			if tt.wantDetails == "" {
				require.Empty(t, csErr.Details)
			} else {
				require.JSONEq(t, tt.wantDetails, string(csErr.Details))
			}

			b, mErr := json.Marshal(err)
			require.NoError(t, mErr)
			var stored changeset.Error
			require.NoError(t, json.Unmarshal(b, &stored))
			require.Equal(t, csErr.Message, stored.Message)
		})
	}
}
