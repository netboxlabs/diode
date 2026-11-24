package postgres

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

// View returns a claimed row in the v_ingestion_logs shape. The claim
// queries read the changed columns from their UPDATE ... RETURNING, so sqlc
// gives them their own row type.
func (r ClaimQueuedIngestionLogsRow) View() VIngestionLog {
	return VIngestionLog{
		ID:                 r.ID,
		ExternalID:         r.ExternalID,
		ObjectType:         r.ObjectType,
		State:              r.State,
		RequestID:          r.RequestID,
		IngestionTs:        r.IngestionTs,
		SourceTs:           r.SourceTs,
		ProducerAppName:    r.ProducerAppName,
		ProducerAppVersion: r.ProducerAppVersion,
		SdkName:            r.SdkName,
		SdkVersion:         r.SdkVersion,
		Entity:             r.Entity,
		Error:              r.Error,
		SourceMetadata:     r.SourceMetadata,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
		EntityHash:         r.EntityHash,
		LastSeen:           r.LastSeen,
		DuplicateCount:     r.DuplicateCount,
		RequeuedFromState:  r.RequeuedFromState,
	}
}

// View returns a claimed row in the v_ingestion_logs shape.
func (r ClaimQueuedForAutoApplyRow) View() VIngestionLog {
	return ClaimQueuedIngestionLogsRow(r).View()
}

// ToProto converts sqlc structure to analogous protobuf
func (log VIngestionLog) ToProto() (*reconcilerpb.IngestionLog, error) {
	entity := &diodepb.Entity{}
	if err := protojson.Unmarshal(log.Entity, entity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}
	var ingestionErr reconcilerpb.DeviationError
	if log.Error != nil {
		if err := LoadDeviationError(log.Error, &ingestionErr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error: %w", err)
		}
	}

	pblog := &reconcilerpb.IngestionLog{
		Id:                 log.ExternalID,
		DataType:           log.ObjectType.String, // backwards compatibility
		ObjectType:         log.ObjectType.String,
		State:              reconcilerpb.State(log.State.Int32),
		RequestId:          log.RequestID.String,
		IngestionTs:        log.IngestionTs.Int64,
		SourceTs:           log.SourceTs.Int64,
		ProducerAppName:    log.ProducerAppName.String,
		ProducerAppVersion: log.ProducerAppVersion.String,
		SdkName:            log.SdkName.String,
		SdkVersion:         log.SdkVersion.String,
		Entity:             entity,
		Error:              &ingestionErr,
	}

	return pblog, nil
}

// LoadDeviationError loads a deviation error from a byte slice
func LoadDeviationError(errorBytes []byte, deviationErr *reconcilerpb.DeviationError) error {
	if len(errorBytes) == 0 {
		return nil
	}

	var changeSetErr changeset.Error
	if err := json.Unmarshal(errorBytes, &changeSetErr); err != nil {
		return fmt.Errorf("failed to unmarshal error: %w", err)
	}
	deviationErr.Message = changeSetErr.Message
	deviationErr.Code = string(changeSetErr.Code)
	deviationErr.Details = changeSetErr.Details

	return nil
}
