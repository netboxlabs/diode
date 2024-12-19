package reconciler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

func retrieveIngestionMetrics(ctx context.Context, repository Repository) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {
	var metrics reconcilerpb.IngestionMetrics

	ingestionLogsPerState, err := repository.CountIngestionLogsPerState(ctx)
	if err != nil {
		return nil, err
	}

	var totalIngestionLogs int32

	for state, count := range ingestionLogsPerState {
		totalIngestionLogs += count
		switch state {
		case reconcilerpb.State_QUEUED:
			metrics.Queued = count
		case reconcilerpb.State_RECONCILED:
			metrics.Reconciled = count
		case reconcilerpb.State_FAILED:
			metrics.Failed = count
		case reconcilerpb.State_NO_CHANGES:
			metrics.NoChanges = count
		}
	}

	metrics.Total = totalIngestionLogs

	return &reconcilerpb.RetrieveIngestionLogsResponse{Metrics: &metrics}, nil
}

func retrieveIngestionLogs(ctx context.Context, logger *slog.Logger, repository Repository, in *reconcilerpb.RetrieveIngestionLogsRequest) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {
	if in.GetOnlyMetrics() {
		logger.Debug("retrieving only ingestion metrics")
		return retrieveIngestionMetrics(ctx, repository)
	}

	pageSize := in.GetPageSize()
	if in.PageSize == nil || pageSize >= 1000 {
		pageSize = 100
	}

	offset := int32(0)
	if in.PageToken != "" {
		decodedPageToken, err := decodeBase64ToInt(in.PageToken)
		if err != nil {
			return nil, err
		}
		offset = decodedPageToken
	}

	logs, err := repository.RetrieveIngestionLogs(ctx, in, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ingestion logs: %w", err)
	}

	var nextPageToken string

	if len(logs) == int(pageSize) {
		offset += int32(len(logs))
		nextPageToken = encodeIntToBase64(offset)
	}

	// Fill metrics
	var metrics reconcilerpb.IngestionMetrics
	total := int32(len(logs))
	if in.State != nil {
		if in.GetState() == reconcilerpb.State_UNSPECIFIED {
			metrics.Total = total
		} else if in.GetState() == reconcilerpb.State_QUEUED {
			metrics.Queued = total
		} else if in.GetState() == reconcilerpb.State_RECONCILED {
			metrics.Reconciled = total
		} else if in.GetState() == reconcilerpb.State_FAILED {
			metrics.Failed = total
		} else if in.GetState() == reconcilerpb.State_NO_CHANGES {
			metrics.NoChanges = total
		}
	} else {
		metrics.Total = total
	}

	return &reconcilerpb.RetrieveIngestionLogsResponse{Logs: logs, Metrics: &metrics, NextPageToken: nextPageToken}, nil
}

func decodeBase64ToInt(encoded string) (int32, error) {
	// Decode the base64 string back to bytes
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return 0, err
	}

	// Convert the byte slice back to int64
	buf := bytes.NewReader(decoded)
	var num int32
	if err := binary.Read(buf, binary.BigEndian, &num); err != nil {
		return 0, err
	}

	return num, nil
}

func encodeIntToBase64(num int32) string {
	// Create a buffer to hold the binary representation
	buf := new(bytes.Buffer)

	// Write the int value as a binary value into the buffer
	if err := binary.Write(buf, binary.BigEndian, num); err != nil {
		fmt.Println("error writing binary:", err)
	}

	// Encode the binary data to base64
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
