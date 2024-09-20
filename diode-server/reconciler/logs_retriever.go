package reconciler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

type extraAttributesWrapper struct {
	ExtraAttributes string `json:"$"`
	IngestionTs     string `json:"ingestion_ts"`
}
type redisLogResult struct {
	ExtraAttributes extraAttributesWrapper `json:"extra_attributes"`
	ID              string                 `json:"id"`
}

type redisLogsResponse struct {
	Results      []redisLogResult `json:"results"`
	TotalResults int32            `json:"total_results"`
}

func convertMapInterface(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		converted := make(map[string]interface{})
		for key, value := range v {
			converted[fmt.Sprintf("%v", key)] = convertMapInterface(value) // Recursive conversion for nested maps
		}
		return converted
	case []interface{}:
		// If the value is a slice, apply conversion recursively to each element
		for i, value := range v {
			v[i] = convertMapInterface(value)
		}
		return v
	default:
		return v
	}
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

func retrieveIngestionMetrics(ctx context.Context, client RedisClient) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {
	pipe := client.Pipeline()

	results := []*redis.Cmd{
		pipe.Do(ctx, "FT.SEARCH", "ingest-entity", "*", "LIMIT", 0, 0),
	}
	for s := reconcilerpb.State_NEW; s <= reconcilerpb.State_NO_CHANGES; s++ {
		stateName, ok := reconcilerpb.State_name[int32(s)]
		if !ok {
			return nil, fmt.Errorf("failed to retrieve ingestion logs: failed to get state name of %d", s)
		}
		results = append(results, pipe.Do(ctx, "FT.SEARCH", "ingest-entity", fmt.Sprintf("@state:%s", stateName), "LIMIT", 0, 0))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to retrieve ingestion logs: %w", err)
	}

	var metrics reconcilerpb.IngestionMetrics

	for q := range results {
		res, err := results[q].Result()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve ingestion logs: %w", err)
		}

		conv := convertMapInterface(res)
		totalRes, ok := conv.(map[string]interface{})["total_results"].(int64)
		if !ok {
			return nil, fmt.Errorf("failed to retrieve ingestion logs: failed to parse total_results")
		}
		total := int32(totalRes)
		if q == int(reconcilerpb.State_NEW) {
			metrics.New = total
		} else if q == int(reconcilerpb.State_RECONCILED) {
			metrics.Reconciled = total
		} else if q == int(reconcilerpb.State_FAILED) {
			metrics.Failed = total
		} else if q == int(reconcilerpb.State_NO_CHANGES) {
			metrics.NoChanges = total
		} else {
			metrics.Total = total
		}
	}
	return &reconcilerpb.RetrieveIngestionLogsResponse{Logs: nil, Metrics: &metrics, NextPageToken: ""}, nil
}

func retrieveIngestionLogs(ctx context.Context, logger *slog.Logger, client RedisClient, in *reconcilerpb.RetrieveIngestionLogsRequest) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {
	if in.GetOnlyMetrics() {
		logger.Debug("retrieving only ingestion metrics")
		return retrieveIngestionMetrics(ctx, client)
	}

	pageSize := in.GetPageSize()
	if in.PageSize == nil || pageSize >= 1000 {
		pageSize = 100 // Default to 100
	}

	query := buildQueryFilter(in)

	// Construct the base FT.SEARCH query
	queryArgs := []interface{}{
		"FT.SEARCH",
		"ingest-entity", // Index name
		query,
	}

	// Apply sorting by id in descending order
	queryArgs = append(queryArgs, "SORTBY", "id", "DESC")

	var err error

	// Apply limit for pagination
	var offset int32
	if in.PageToken != "" {
		offset, err = decodeBase64ToInt(in.PageToken)
		if err != nil {
			logger.Warn("error decoding page token", "error", err)
		}
	}
	queryArgs = append(queryArgs, "LIMIT", offset, pageSize)

	logger.Debug("retrieving ingestion logs", "query", queryArgs)

	// Execute the query using Redis
	result, err := client.Do(ctx, queryArgs...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ingestion logs: %w", err)
	}

	res := convertMapInterface(result)

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("error marshaling ingestion logs: %w", err)
	}

	var response redisLogsResponse

	// Unmarshal the result into the struct
	if err = json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	logs := make([]*reconcilerpb.IngestionLog, 0)

	for _, logsResult := range response.Results {
		ingestionLog := &reconcilerpb.IngestionLog{}
		if err := protojson.Unmarshal([]byte(logsResult.ExtraAttributes.ExtraAttributes), ingestionLog); err != nil {
			return nil, fmt.Errorf("error parsing ExtraAttributes JSON: %v", err)
		}

		logs = append(logs, ingestionLog)
	}

	var nextPageToken string

	if len(logs) == int(pageSize) {
		offset += int32(len(logs))
		nextPageToken = encodeIntToBase64(offset)
	}

	// Fill metrics
	var metrics reconcilerpb.IngestionMetrics
	if in.State != nil {
		if in.GetState() == reconcilerpb.State_UNSPECIFIED {
			metrics.Total = response.TotalResults
		} else if in.GetState() == reconcilerpb.State_NEW {
			metrics.New = response.TotalResults
		} else if in.GetState() == reconcilerpb.State_RECONCILED {
			metrics.Reconciled = response.TotalResults
		} else if in.GetState() == reconcilerpb.State_FAILED {
			metrics.Failed = response.TotalResults
		} else if in.GetState() == reconcilerpb.State_NO_CHANGES {
			metrics.NoChanges = response.TotalResults
		}
	} else {
		metrics.Total = response.TotalResults
	}

	return &reconcilerpb.RetrieveIngestionLogsResponse{Logs: logs, Metrics: &metrics, NextPageToken: nextPageToken}, nil
}

func buildQueryFilter(req *reconcilerpb.RetrieveIngestionLogsRequest) string {
	queryFilter := "*"

	// apply optional filters for ingestion timestamps (start and end)
	if req.GetIngestionTsStart() > 0 || req.GetIngestionTsEnd() > 0 {
		ingestionTsFilter := fmt.Sprintf("@ingestion_ts:[%d inf]", req.GetIngestionTsStart())

		if req.GetIngestionTsEnd() > 0 {
			ingestionTsFilter = fmt.Sprintf("@ingestion_ts:[%d %d]", req.GetIngestionTsStart(), req.GetIngestionTsEnd())
		}

		queryFilter = ingestionTsFilter
	}

	// apply optional filters for ingestion state
	if req.State != nil {
		stateFilter := fmt.Sprintf("@state:%s", req.GetState().String())
		if queryFilter == "*" {
			queryFilter = stateFilter
		} else {
			queryFilter = fmt.Sprintf("%s %s", queryFilter, stateFilter)
		}
	}

	if req.GetDataType() != "" {
		dataTypeFilter := fmt.Sprintf("@data_type:%s", req.GetDataType())
		if queryFilter == "*" {
			queryFilter = dataTypeFilter
		} else {
			queryFilter = fmt.Sprintf("%s %s", queryFilter, dataTypeFilter)
		}
	}

	return queryFilter
}
