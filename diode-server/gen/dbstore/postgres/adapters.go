package postgres

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

// ToProto converts sqlc structure to analogous protobuf
func (log IngestionLog) ToProto() (*reconcilerpb.IngestionLog, error) {
	entity := &diodepb.Entity{}
	if err := protojson.Unmarshal(log.Entity, entity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}
	var ingestionErr reconcilerpb.IngestionError
	if log.Error != nil {
		if err := protojson.Unmarshal(log.Error, &ingestionErr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error: %w", err)
		}
	}

	pblog := &reconcilerpb.IngestionLog{
		Id:                 log.ExternalID,
		DataType:           log.DataType.String,
		State:              reconcilerpb.State(log.State.Int32),
		RequestId:          log.RequestID.String,
		IngestionTs:        log.IngestionTs.Int64,
		ProducerAppName:    log.ProducerAppName.String,
		ProducerAppVersion: log.ProducerAppVersion.String,
		SdkName:            log.SdkName.String,
		SdkVersion:         log.SdkVersion.String,
		Entity:             entity,
		Error:              &ingestionErr,
	}

	return pblog, nil
}
