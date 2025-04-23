package reconciler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	mr "github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func TestRetrieveDeviations(t *testing.T) {
	tests := []struct {
		name            string
		deviations      []*reconcilerpb.Deviation
		req             *reconcilerpb.RetrieveDeviationsRequest
		resp            *reconcilerpb.RetrieveDeviationsResponse
		changeAfterJSON []byte
		repoError       error
		wantErr         error
	}{
		{
			name: "deviations found",
			req:  &reconcilerpb.RetrieveDeviationsRequest{},
			resp: &reconcilerpb.RetrieveDeviationsResponse{
				Deviations: []*reconcilerpb.Deviation{
					{
						Id:         "deviation-id",
						Name:       "Platform X modified",
						Source:     "orb-agent",
						State:      reconcilerpb.State_APPLIED,
						ObjectType: "dcim.platform",
						BranchId:   strPtr("branch-id"),
						IngestedEntity: &diodepb.Entity{
							Entity: &diodepb.Entity_Platform{
								Platform: &diodepb.Platform{
									Name:        "X",
									Description: strPtr("Example description"),
								},
							},
						},
						Error: nil,
						Changes: []*reconcilerpb.Change{
							{
								Id:                 "change-id",
								ObjectType:         "dcim.platform",
								ChangeType:         changeset.ChangeTypeUpdate,
								ObjectPrimaryValue: "X",
								Before:             []byte(`{"id": 1, "name": "X", "description": ""}`),
								After:              []byte(`{"id": 1, "name": "X", "description": "Example description"}`),
							},
						},
						IngestionTs:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
						LastUpdateTs: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
						SourceTs:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
					},
				},
			},
			deviations: []*reconcilerpb.Deviation{
				{
					Id:         "deviation-id",
					Name:       "Platform X modified",
					Source:     "orb-agent",
					State:      reconcilerpb.State_APPLIED,
					ObjectType: "dcim.platform",
					BranchId:   strPtr("branch-id"),
					IngestedEntity: &diodepb.Entity{
						Entity: &diodepb.Entity_Platform{
							Platform: &diodepb.Platform{
								Name:        "X",
								Description: strPtr("Example description"),
							},
						},
					},
					Error: nil,
					Changes: []*reconcilerpb.Change{
						{
							Id:                 "change-id",
							ObjectType:         "dcim.platform",
							ChangeType:         changeset.ChangeTypeUpdate,
							ObjectPrimaryValue: "X",
							Before:             []byte(`{"id": 1, "name": "X", "description": ""}`),
							After:              []byte(`{"id": 1, "name": "X", "description": "Example description"}`),
						},
					},
					IngestionTs:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
					LastUpdateTs: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
					SourceTs:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
				},
			},
			changeAfterJSON: []byte(`{"id": 1, "name": "X", "description": "Example description"}`),
			repoError:       nil,
			wantErr:         nil,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			mockRedisClient := mr.NewRedisClient(t)
			mockRepository := mr.NewRepository(t)
			server := &Server{
				redisClient: mockRedisClient,
				logger:      logger,
				repository:  mockRepository,
			}

			mockRepository.On("RetrieveDeviations", ctx, tt.req, mock.Anything, mock.Anything).Return(tt.deviations, tt.repoError)

			response, err := server.RetrieveDeviations(ctx, tt.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Nil(t, response)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)

				for i, deviation := range response.Deviations {
					assert.Equal(t, tt.deviations[i].Id, deviation.Id)
					assert.Equal(t, tt.deviations[i].Name, deviation.Name)
					assert.Equal(t, tt.deviations[i].Source, deviation.Source)
					assert.Equal(t, tt.deviations[i].State, deviation.State)
					assert.Equal(t, tt.deviations[i].ObjectType, deviation.ObjectType)
					assert.Equal(t, tt.deviations[i].BranchId, deviation.BranchId)
					assert.Equal(t, tt.deviations[i].IngestedEntity.String(), deviation.IngestedEntity.String())
					assert.Equal(t, tt.deviations[i].Error, deviation.Error)
					assert.Equal(t, len(tt.deviations[i].Changes), len(deviation.Changes))
					assert.Equal(t, tt.deviations[i].Changes[0].Before, deviation.Changes[0].Before)
					assert.Equal(t, tt.deviations[i].Changes[0].After, deviation.Changes[0].After)
					assert.Equal(t, tt.changeAfterJSON, deviation.Changes[0].After)
					require.NotZero(t, deviation.IngestionTs)
					require.NotZero(t, deviation.LastUpdateTs)
					require.NotZero(t, deviation.SourceTs)
					assert.Equal(t, tt.deviations[i].IngestionTs, deviation.IngestionTs)
					assert.Equal(t, tt.deviations[i].LastUpdateTs, deviation.LastUpdateTs)
					assert.Equal(t, tt.resp.Deviations[i].SourceTs, deviation.SourceTs)
				}
			}
			mockRepository.AssertExpectations(t)
		})
	}
}

func TestRetrieveDeviationByID(t *testing.T) {
	tests := []struct {
		name            string
		deviation       *reconcilerpb.Deviation
		req             *reconcilerpb.RetrieveDeviationByIDRequest
		resp            *reconcilerpb.RetrieveDeviationByIDResponse
		changeAfterJSON []byte
		repoError       error
		wantErr         error
	}{
		{
			name: "deviation found",
			req:  &reconcilerpb.RetrieveDeviationByIDRequest{Id: "deviation-id"},
			resp: &reconcilerpb.RetrieveDeviationByIDResponse{
				Deviation: &reconcilerpb.Deviation{
					Id:         "deviation-id",
					Name:       "Platform X modified",
					Source:     "orb-agent",
					State:      reconcilerpb.State_APPLIED,
					ObjectType: "dcim.platform",
					BranchId:   strPtr("branch-id"),
					IngestedEntity: &diodepb.Entity{
						Entity: &diodepb.Entity_Platform{
							Platform: &diodepb.Platform{
								Name:        "X",
								Description: strPtr("Example description"),
							},
						},
					},
					Error: nil,
					Changes: []*reconcilerpb.Change{
						{
							Id:                 "change-id",
							ObjectType:         "dcim.platform",
							ChangeType:         changeset.ChangeTypeUpdate,
							ObjectPrimaryValue: "X",
							Before:             []byte(`{"id": 1, "name": "X", "description": ""}`),
							After:              []byte(`{"id": 1, "name": "X", "description": "Example description"}`),
						},
					},
					IngestionTs:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
					LastUpdateTs: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
				},
			},
			deviation: &reconcilerpb.Deviation{
				Id:         "deviation-id",
				Name:       "Platform X modified",
				Source:     "orb-agent",
				State:      reconcilerpb.State_APPLIED,
				ObjectType: "dcim.platform",
				BranchId:   strPtr("branch-id"),
				IngestedEntity: &diodepb.Entity{
					Entity: &diodepb.Entity_Platform{
						Platform: &diodepb.Platform{
							Name:        "X",
							Description: strPtr("Example description"),
						},
					},
				},
				Error: nil,
				Changes: []*reconcilerpb.Change{
					{
						Id:                 "change-id",
						ObjectType:         "dcim.platform",
						ChangeType:         changeset.ChangeTypeUpdate,
						ObjectPrimaryValue: "X",
						Before:             []byte(`{"id": 1, "name": "X", "description": ""}`),
						After:              []byte(`{"id": 1, "name": "X", "description": "Example description"}`),
					},
				},
				IngestionTs:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
				LastUpdateTs: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			},
			changeAfterJSON: []byte(`{"id": 1, "name": "X", "description": "Example description"}`),
			repoError:       nil,
			wantErr:         nil,
		},
		{
			name:      "deviation not found",
			req:       &reconcilerpb.RetrieveDeviationByIDRequest{Id: "deviation-id"},
			resp:      nil,
			deviation: nil,
			repoError: errors.New("no rows in result set"),
			wantErr:   errors.New("failed to retrieve deviation: no rows in result set"),
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			mockRedisClient := mr.NewRedisClient(t)
			mockRepository := mr.NewRepository(t)
			server := &Server{
				redisClient: mockRedisClient,
				logger:      logger,
				repository:  mockRepository,
			}

			mockRepository.On("RetrieveDeviationByID", ctx, tt.req.GetId()).Return(tt.deviation, tt.repoError)

			response, err := server.RetrieveDeviationByID(ctx, tt.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Nil(t, response)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)

				assert.Equal(t, tt.resp.Deviation.Id, response.Deviation.Id)
				assert.Equal(t, tt.resp.Deviation.Name, response.Deviation.Name)
				assert.Equal(t, tt.resp.Deviation.Source, response.Deviation.Source)
				assert.Equal(t, tt.resp.Deviation.State, response.Deviation.State)
				assert.Equal(t, tt.resp.Deviation.ObjectType, response.Deviation.ObjectType)
				assert.Equal(t, tt.resp.Deviation.BranchId, response.Deviation.BranchId)
				assert.Equal(t, tt.resp.Deviation.IngestedEntity.String(), response.Deviation.IngestedEntity.String())
				assert.Equal(t, tt.resp.Deviation.Error, response.Deviation.Error)
				assert.Equal(t, len(tt.resp.Deviation.Changes), len(response.Deviation.Changes))
				assert.Equal(t, tt.resp.Deviation.Changes[0].Before, response.Deviation.Changes[0].Before)
				assert.Equal(t, tt.resp.Deviation.Changes[0].After, response.Deviation.Changes[0].After)
				assert.Equal(t, tt.changeAfterJSON, response.Deviation.Changes[0].After)
				require.NotZero(t, response.Deviation.IngestionTs)
				require.NotZero(t, response.Deviation.LastUpdateTs)
				assert.Equal(t, tt.resp.Deviation.IngestionTs, response.Deviation.IngestionTs)
				assert.Equal(t, tt.resp.Deviation.LastUpdateTs, response.Deviation.LastUpdateTs)
			}
			mockRepository.AssertExpectations(t)
		})
	}
}
