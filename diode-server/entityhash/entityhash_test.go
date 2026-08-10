package entityhash_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	diodepb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

func TestEntityFingerprinter_GenerateEntityHash(t *testing.T) {
	fingerprinter := entityhash.NewEntityFingerprinter()

	tests := []struct {
		name           string
		entity         *diodepb.Entity
		expectError    bool
		errorContains  string
		expectedLength int
	}{
		{
			name:          "nil entity returns error",
			entity:        nil,
			expectError:   true,
			errorContains: "entity cannot be nil",
		},
		{
			name: "entity with nil content returns error",
			entity: &diodepb.Entity{
				Timestamp: timestamppb.Now(),
				Entity:    nil,
			},
			expectError:   true,
			errorContains: "entity content cannot be nil",
		},
		{
			name: "simple device entity produces valid hash",
			entity: &diodepb.Entity{
				Timestamp: timestamppb.Now(),
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("test-device"),
					},
				},
			},
			expectError:    false,
			expectedLength: 64, // SHA256 hex length
		},
		{
			name: "simple site entity produces valid hash",
			entity: &diodepb.Entity{
				Timestamp: timestamppb.Now(),
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{
						Name: "test-site",
					},
				},
			},
			expectError:    false,
			expectedLength: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := fingerprinter.GenerateEntityHash(tt.entity)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hash)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, hash, tt.expectedLength)
				assert.Regexp(t, "^[a-f0-9]{64}$", hash, "hash should be lowercase hex")
			}
		})
	}
}

func TestEntityFingerprinter_BasicEntities(t *testing.T) {
	fingerprinter := entityhash.NewEntityFingerprinter()

	timestamp := timestamppb.Now()
	tests := []struct {
		name        string
		entity1     *diodepb.Entity
		entity2     *diodepb.Entity
		shouldMatch bool
	}{
		{
			name: "identical devices produce same hash",
			entity1: &diodepb.Entity{
				Timestamp: timestamp,
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name:   strPtr("device1"),
						Serial: strPtr("serial123"),
					},
				},
			},
			entity2: &diodepb.Entity{
				// Different timestamp is ignoreds
				Timestamp: timestamppb.New(timestamp.AsTime().Add(10 * time.Second)),
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name:   strPtr("device1"),
						Serial: strPtr("serial123"),
					},
				},
			},
			shouldMatch: true,
		},
		{
			name: "different device names produce different hashes",
			entity1: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("device1"),
					},
				},
			},
			entity2: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("device2"),
					},
				},
			},
			shouldMatch: false,
		},
		{
			name: "same device with different metadata produces same hash",
			entity1: &diodepb.Entity{
				Timestamp: timestamp,
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name:   strPtr("device1"),
						Serial: strPtr("serial123"),
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"run_id": structpb.NewStringValue("run-1"),
							},
						},
					},
				},
			},
			entity2: &diodepb.Entity{
				Timestamp: timestamp,
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name:   strPtr("device1"),
						Serial: strPtr("serial123"),
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"run_id": structpb.NewStringValue("run-2"),
							},
						},
					},
				},
			},
			shouldMatch: true,
		},
		{
			name: "same device with and without metadata produces same hash",
			entity1: &diodepb.Entity{
				Timestamp: timestamp,
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name:   strPtr("device1"),
						Serial: strPtr("serial123"),
					},
				},
			},
			entity2: &diodepb.Entity{
				Timestamp: timestamp,
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name:   strPtr("device1"),
						Serial: strPtr("serial123"),
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"run_id": structpb.NewStringValue("run-1"),
							},
						},
					},
				},
			},
			shouldMatch: true,
		},
		{
			name: "same site with different metadata produces same hash",
			entity1: &diodepb.Entity{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{
						Name: "site1",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"job_name": structpb.NewStringValue("job-a"),
							},
						},
					},
				},
			},
			entity2: &diodepb.Entity{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{
						Name: "site1",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"job_name": structpb.NewStringValue("job-b"),
							},
						},
					},
				},
			},
			shouldMatch: true,
		},
		{
			name: "device with nested entity metadata produces same hash",
			entity1: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("device1"),
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"run_id": structpb.NewStringValue("run-1"),
							},
						},
						Site: &diodepb.Site{
							Name: "site1",
							Metadata: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"run_id": structpb.NewStringValue("run-1"),
								},
							},
						},
					},
				},
			},
			entity2: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("device1"),
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"run_id": structpb.NewStringValue("run-2"),
							},
						},
						Site: &diodepb.Site{
							Name: "site1",
							Metadata: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"run_id": structpb.NewStringValue("run-2"),
								},
							},
						},
					},
				},
			},
			shouldMatch: true,
		},
		{
			name: "different entity types produce different hashes",
			entity1: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("test"),
					},
				},
			},
			entity2: &diodepb.Entity{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{
						Name: "test",
					},
				},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err1 := fingerprinter.GenerateEntityHash(tt.entity1)
			hash2, err2 := fingerprinter.GenerateEntityHash(tt.entity2)

			require.NoError(t, err1)
			require.NoError(t, err2)

			if tt.shouldMatch {
				assert.Equal(t, hash1, hash2, "hashes should be identical")
			} else {
				assert.NotEqual(t, hash1, hash2, "hashes should be different")
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
