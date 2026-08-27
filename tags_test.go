package lunchmoney

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        *CreateTag
		statusCode int
		response   string
		wantBody   map[string]any
		wantErr    string
	}{
		{
			name:       "name only",
			tag:        &CreateTag{Name: "Date Night"},
			statusCode: http.StatusCreated,
			response:   `{"id": 94350, "name": "Date Night", "description": null, "archived": false, "archived_at": null}`,
			wantBody:   map[string]any{"name": "Date Night"},
		},
		{
			name:       "with description and colors",
			tag:        &CreateTag{Name: "Date Night", Description: "dinners", TextColor: "333", BackgroundColor: "FFE7D4", Archived: true},
			statusCode: http.StatusCreated,
			response:   `{"id": 94351, "name": "Date Night", "description": "dinners", "archived": true}`,
			wantBody: map[string]any{
				"name":             "Date Night",
				"description":      "dinners",
				"text_color":       "333",
				"background_color": "FFE7D4",
				"archived":         true,
			},
		},
		{
			name:    "name is required",
			tag:     &CreateTag{},
			wantErr: "Name",
		},
		{
			name:    "name is too long",
			tag:     &CreateTag{Name: strings.Repeat("a", 101)},
			wantErr: "Name",
		},
		{
			name:       "duplicate name",
			tag:        &CreateTag{Name: "Date Night"},
			statusCode: http.StatusBadRequest,
			response:   `{"message": "Invalid Request Body", "errors": [{"errMsg": "Tag with name 'Date Night' already exists"}]}`,
			wantErr:    "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/tags", r.URL.Path)
				require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			tag, err := testClient(t, server).CreateTag(context.Background(), tt.tag)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.tag.Name, tag.Name)
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}

func TestUpdateTag(t *testing.T) {
	name := "Updated Tag Name"
	archived := true

	tests := []struct {
		name       string
		tag        *UpdateTag
		statusCode int
		response   string
		wantBody   map[string]any
		wantErr    string
	}{
		{
			name:       "name only",
			tag:        &UpdateTag{Name: &name},
			statusCode: http.StatusOK,
			response:   `{"id": 94319, "name": "Updated Tag Name", "description": null}`,
			// The fields left nil stay out of the request body entirely.
			wantBody: map[string]any{"name": "Updated Tag Name"},
		},
		{
			name:       "archive",
			tag:        &UpdateTag{Archived: &archived},
			statusCode: http.StatusOK,
			response:   `{"id": 94319, "name": "Date Night", "archived": true}`,
			wantBody:   map[string]any{"archived": true},
		},
		{
			name:    "name cannot be empty",
			tag:     &UpdateTag{Name: new(string)},
			wantErr: "Name",
		},
		{
			name:       "not found",
			tag:        &UpdateTag{Name: &name},
			statusCode: http.StatusNotFound,
			response:   `{"message": "Not Found", "errors": [{"errMsg": "There is no tag with the id: 543210."}]}`,
			wantErr:    "no tag with the id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/tags/94319", r.URL.Path)
				require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			tag, err := testClient(t, server).UpdateTag(context.Background(), 94319, tt.tag)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, int64(94319), tag.ID)
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}

func TestDeleteTag(t *testing.T) {
	tests := []struct {
		name       string
		force      bool
		statusCode int
		response   string
		wantForce  string
		wantErr    string
	}{
		{
			name:       "deleted",
			statusCode: http.StatusNoContent,
		},
		{
			name:       "forced",
			force:      true,
			statusCode: http.StatusNoContent,
			wantForce:  "true",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   `{"message": "Not Found", "errors": [{"errMsg": "There is no tag with the id: 543210."}]}`,
			wantErr:    "no tag with the id",
		},
		{
			name:       "still in use",
			statusCode: http.StatusUnprocessableEntity,
			response:   `{"tag_name": "Tag to be Deleted", "dependents": {"rules": 1, "transactions": 10}}`,
			wantErr:    "used by 1 rules and 10 transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "/tags/94319", r.URL.Path)
				assert.Equal(t, tt.wantForce, r.URL.Query().Get("force"))

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			err := testClient(t, server).DeleteTag(context.Background(), 94319, tt.force)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDeleteTagInUseIsInspectable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, err := w.Write([]byte(`{"tag_name": "Tag to be Deleted", "dependents": {"rules": 1, "transactions": 10}}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	err := testClient(t, server).DeleteTag(context.Background(), 94317, false)
	require.Error(t, err)

	var inUse *TagInUseError
	require.True(t, errors.As(err, &inUse))
	assert.Equal(t, "Tag to be Deleted", inUse.TagName)
	assert.Equal(t, int64(1), inUse.Dependents.Rules)
	assert.Equal(t, int64(10), inUse.Dependents.Transactions)

	// The API error the dependents arrived as stays reachable underneath.
	var apiErr *ErrorResponse
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.StatusCode)
}
