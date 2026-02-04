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

func retrieveDeviations(ctx context.Context, _ *slog.Logger, repository Repository, req *reconcilerpb.RetrieveDeviationsRequest) (*reconcilerpb.RetrieveDeviationsResponse, error) {
	pageSize := req.GetPageSize()
	if req.PageSize == nil || pageSize >= 1000 {
		pageSize = 100
	}

	offset := int32(0)
	if req.PageToken != "" {
		decodedPageToken, err := decodePageToken(req.PageToken)
		if err != nil {
			return nil, err
		}
		offset = decodedPageToken
	}

	deviations, err := repository.RetrieveDeviations(ctx, req, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve deviations: %w", err)
	}

	var nextPageToken string

	if len(deviations) == int(pageSize) {
		offset += int32(len(deviations))
		nextPageToken = encodePageNumber(offset)
	}

	return &reconcilerpb.RetrieveDeviationsResponse{Deviations: deviations, NextPageToken: nextPageToken}, nil
}

func retrieveDeviationByID(ctx context.Context, _ *slog.Logger, repository Repository, req *reconcilerpb.RetrieveDeviationByIDRequest) (*reconcilerpb.RetrieveDeviationByIDResponse, error) {
	deviation, err := repository.RetrieveDeviationByID(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve deviation: %w", err)
	}

	return &reconcilerpb.RetrieveDeviationByIDResponse{Deviation: deviation}, nil
}

func listResultsByJob(ctx context.Context, _ *slog.Logger, repository Repository, req *reconcilerpb.ListResultsByJobRequest) (*reconcilerpb.ListResultsByJobResponse, error) {
	pageSize := req.GetPageSize()
	if req.PageSize == nil || pageSize >= 1000 {
		pageSize = 100
	}

	offset := int32(0)
	if req.PageToken != "" {
		decodedPageToken, err := decodePageToken(req.PageToken)
		if err != nil {
			return nil, err
		}
		offset = decodedPageToken
	}

	deviations, err := repository.ListResultsByJob(ctx, req, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list results by job: %w", err)
	}

	var nextPageToken string

	if len(deviations) == int(pageSize) {
		offset += int32(len(deviations))
		nextPageToken = encodePageNumber(offset)
	}

	return &reconcilerpb.ListResultsByJobResponse{Deviations: deviations, NextPageToken: nextPageToken}, nil
}

func decodePageToken(encoded string) (int32, error) {
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

func encodePageNumber(num int32) string {
	// Create a buffer to hold the binary representation
	buf := new(bytes.Buffer)

	// Write the int value as a binary value into the buffer
	if err := binary.Write(buf, binary.BigEndian, num); err != nil {
		fmt.Println("error writing binary:", err)
	}

	// Encode the binary data to base64
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
