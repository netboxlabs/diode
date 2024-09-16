package reconciler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/redis/go-redis/v9"

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

func encodeInt64ToBase64(num int64) string {
	// Create a buffer to hold the binary representation
	buf := new(bytes.Buffer)

	// Write the int64 value as a binary value into the buffer
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		fmt.Println("Error writing binary:", err)
	}

	// Encode the binary data to base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded
}

func decodeBase64ToInt64(encoded string) (int64, error) {
	// Decode the base64 string back to bytes
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return 0, err
	}

	// Convert the byte slice back to int64
	var num int64
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.BigEndian, &num)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func retrieveIngestionStatsSummary(ctx context.Context, client RedisClient) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {

	pipe := client.Pipeline()

	results := make([]*redis.Cmd, 0)
	results = append(results, pipe.Do(ctx, "FT.SEARCH", "ingest-entity", "*", "LIMIT", 0, 0))
	for i := 1; i < 5; i++ {
		results = append(results, pipe.Do(ctx, "FT.SEARCH", "ingest-entity", fmt.Sprintf("@state:[%d %d]", i, i), "LIMIT", 0, 0))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ingestion logs: %w", err)
	}

	var stats reconcilerpb.Stats

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
			stats.New = &total
		} else if q == int(reconcilerpb.State_RECONCILED) {
			stats.Reconciled = &total
		} else if q == int(reconcilerpb.State_FAILED) {
			stats.Failed = &total
		} else if q == int(reconcilerpb.State_NO_CHANGES) {
			stats.NoChanges = &total
		} else {
			stats.Total = &total
		}
	}
	return &reconcilerpb.RetrieveIngestionLogsResponse{Logs: nil, Stats: &stats, NextPageToken: ""}, nil
}

func retrieveIngestionLogs(ctx context.Context, logger *slog.Logger, client RedisClient, in *reconcilerpb.RetrieveIngestionLogsRequest) (*reconcilerpb.RetrieveIngestionLogsResponse, error) {
	if in.GetSummary() {
		logger.Debug("retrieving ingestion logs summary")
		return retrieveIngestionStatsSummary(ctx, client)
	}

	pageSize := in.GetPageSize()
	if in.PageSize == nil {
		pageSize = 100 // Default to 100
	}

	var err error
	var ingestionTs int64

	//Check start TS filter
	var startTs int64
	if in.GetIngestionTsStart() != 0 {
		startTs = in.GetIngestionTsStart()
	}
	query := fmt.Sprintf("@ingestion_ts:[%d inf]", startTs)

	if in.PageToken != "" {
		ingestionTs, err = decodeBase64ToInt64(in.PageToken)
		if err != nil {
			return nil, fmt.Errorf("error decoding page token: %w", err)
		}
		query = fmt.Sprintf("@ingestion_ts:[%d %d]", startTs, ingestionTs)
	}

	// Construct the base FT.SEARCH query
	queryArgs := []interface{}{
		"FT.SEARCH",
		"ingest-entity", // Index name
		query,
	}

	queryIndex := len(queryArgs) - 1

	// Apply optional filters
	if in.State != nil {
		stateFilter := fmt.Sprintf("@state:[%d %d]", in.GetState(), in.GetState())
		queryArgs[queryIndex] = fmt.Sprintf("%s %s", queryArgs[queryIndex], stateFilter)
	}

	if in.GetDataType() != "" {
		dataType := fmt.Sprintf("@data_type:%s", in.GetDataType())
		queryArgs[queryIndex] = fmt.Sprintf("%s %s", queryArgs[queryIndex], dataType)
	}

	// Apply sorting by ingestion_ts in descending order
	queryArgs = append(queryArgs, "SORTBY", "ingestion_ts", "DESC")

	// Apply limit for pagination
	queryArgs = append(queryArgs, "LIMIT", 0, pageSize)

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

		ingestionTs, err = strconv.ParseInt(logsResult.ExtraAttributes.IngestionTs, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error converting ingestion timestamp: %w", err)
		}
	}

	// Fill stats
	var stats reconcilerpb.Stats
	if in.State != nil {
		if in.GetState() == reconcilerpb.State_UNSPECIFIED {
			stats.Total = &response.TotalResults
		} else if in.GetState() == reconcilerpb.State_NEW {
			stats.New = &response.TotalResults
		} else if in.GetState() == reconcilerpb.State_RECONCILED {
			stats.Reconciled = &response.TotalResults
		} else if in.GetState() == reconcilerpb.State_FAILED {
			stats.Failed = &response.TotalResults
		} else if in.GetState() == reconcilerpb.State_NO_CHANGES {
			stats.NoChanges = &response.TotalResults
		}
	} else {
		stats.Total = &response.TotalResults
	}

	return &reconcilerpb.RetrieveIngestionLogsResponse{Logs: logs, Stats: &stats, NextPageToken: encodeInt64ToBase64(ingestionTs)}, nil
}
