package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rhajizada/signum/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadgeModelsJSON(t *testing.T) {
	timestamp := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	subject := "build"

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "marshal create badge response",
			run: func(t *testing.T) {
				payload := models.CreateBadgeResponse{
					Badge: models.Badge{
						ID:        "badge-id",
						Subject:   "build",
						Status:    "passing",
						Color:     "green",
						Style:     "flat",
						CreatedAt: timestamp,
						UpdatedAt: timestamp,
					},
					Token: "token-value",
				}

				body, err := json.Marshal(payload)
				require.NoError(t, err)
				assert.Contains(t, string(body), `"id":"badge-id"`)
				assert.Contains(t, string(body), `"token":"token-value"`)
				assert.Contains(t, string(body), `"subject":"build"`)
			},
		},
		{
			name: "unmarshal patch badge request",
			run: func(t *testing.T) {
				var payload models.PatchBadgeRequest
				err := json.Unmarshal([]byte(`{"subject":"build","status":"passing"}`), &payload)
				require.NoError(t, err)
				require.NotNil(t, payload.Subject)
				require.NotNil(t, payload.Status)
				assert.Equal(t, subject, *payload.Subject)
				assert.Equal(t, "passing", *payload.Status)
				assert.Nil(t, payload.Color)
				assert.Nil(t, payload.Style)
			},
		},
		{
			name: "unmarshal create badge request",
			run: func(t *testing.T) {
				var payload models.CreateBadgeRequest
				err := json.Unmarshal(
					[]byte(`{"subject":"build","status":"passing","color":"green","style":"flat"}`),
					&payload,
				)
				require.NoError(t, err)
				assert.Equal(t, "build", payload.Subject)
				assert.Equal(t, "passing", payload.Status)
				assert.Equal(t, "green", payload.Color)
				assert.Equal(t, "flat", payload.Style)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}
