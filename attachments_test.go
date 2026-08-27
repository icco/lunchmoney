package lunchmoney

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachFileToTransaction(t *testing.T) {
	// The 201 example from the spec. It carries a source field the
	// transactionAttachmentObject schema does not declare, which decoding
	// ignores.
	created := `{
		"id": 1234567890,
		"uploaded_by": 1,
		"name": "receipt.png",
		"type": "image/png",
		"size": 4330,
		"notes": null,
		"source": "api",
		"created_at": "2025-06-11T22:33:20.294Z"
	}`

	tests := []struct {
		name        string
		req         *AttachFileRequest
		statusCode  int
		response    string
		wantName    string
		wantType    string
		wantContent string
		wantErr     string
	}{
		{
			name: "png with notes",
			req: &AttachFileRequest{
				Name:        "receipt.png",
				File:        strings.NewReader("not really a png"),
				ContentType: "image/png",
				Notes:       "Test file attachment",
			},
			statusCode:  http.StatusCreated,
			response:    created,
			wantName:    "receipt.png",
			wantType:    "image/png",
			wantContent: "not really a png",
		},
		{
			name: "pdf without notes",
			req: &AttachFileRequest{
				Name:        "statement.pdf",
				File:        strings.NewReader("%PDF-1.4"),
				ContentType: "application/pdf",
			},
			statusCode:  http.StatusCreated,
			response:    created,
			wantName:    "statement.pdf",
			wantType:    "application/pdf",
			wantContent: "%PDF-1.4",
		},
		{
			name: "filename with a quote in it",
			req: &AttachFileRequest{
				Name:        `we"ird.png`,
				File:        strings.NewReader("not really a png"),
				ContentType: "image/png",
			},
			statusCode:  http.StatusCreated,
			response:    created,
			wantName:    `we"ird.png`,
			wantType:    "image/png",
			wantContent: "not really a png",
		},
		{
			name: "file type the api refuses",
			req: &AttachFileRequest{
				Name:        "archive.zip",
				File:        strings.NewReader("PK"),
				ContentType: "application/zip",
			},
			statusCode:  http.StatusBadRequest,
			response:    `{"message": "Invalid Request Body", "errors": [{"errMsg": "File type application/zip not allowed. Allowed types are: image/jpeg, image/png, application/pdf, image/heic, image/heif"}]}`,
			wantName:    "archive.zip",
			wantType:    "application/zip",
			wantContent: "PK",
			wantErr:     "Invalid Request Body: File type application/zip not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/transactions/42/attachments", r.URL.Path)
				assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;"))

				require.NoError(t, r.ParseMultipartForm(1<<20))

				// The spec names the file field "file"; the server has to see a
				// well formed part under that name, not just the right header.
				file, header, err := r.FormFile("file")
				require.NoError(t, err)
				defer func() { require.NoError(t, file.Close()) }()

				assert.Equal(t, tt.wantName, header.Filename)
				assert.Equal(t, tt.wantType, header.Header.Get("Content-Type"))

				content, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.Equal(t, tt.wantContent, string(content))

				assert.Equal(t, tt.req.Notes, r.FormValue("notes"))

				w.WriteHeader(tt.statusCode)
				_, err = w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).AttachFileToTransaction(context.Background(), 42, tt.req)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Contains(t, err.Error(), "attach file to transaction 42")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, int64(1234567890), got.ID)
			assert.Equal(t, int64(1), got.UploadedBy)
			assert.Equal(t, "receipt.png", got.Name)
			assert.Equal(t, "image/png", got.Type)
			assert.Equal(t, int64(4330), got.Size)
			assert.Empty(t, got.Notes)
			assert.Equal(t, time.Date(2025, 6, 11, 22, 33, 20, 294000000, time.UTC), got.CreatedAt)
		})
	}
}

func TestAttachFileToTransactionValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     *AttachFileRequest
		wantErr string
	}{
		{
			name:    "name is required",
			req:     &AttachFileRequest{File: strings.NewReader("x")},
			wantErr: "Name",
		},
		{
			name:    "file is required",
			req:     &AttachFileRequest{Name: "receipt.png"},
			wantErr: "File",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusCreated)
			}))
			defer server.Close()

			_, err := testClient(t, server).AttachFileToTransaction(context.Background(), 42, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.False(t, called, "an invalid request should not reach the API")
		})
	}
}

func TestAttachFileRequestContentType(t *testing.T) {
	tests := []struct {
		name string
		req  AttachFileRequest
		want string
	}{
		{
			name: "explicit type wins",
			req:  AttachFileRequest{Name: "receipt.png", ContentType: "image/heic"},
			want: "image/heic",
		},
		{
			// Go's extension table is seeded from the host's mime files, so an
			// extension it cannot place has to fall back rather than send "".
			name: "unknown extension falls back",
			req:  AttachFileRequest{Name: "receipt.notarealextension"},
			want: "application/octet-stream",
		},
		{
			name: "no extension falls back",
			req:  AttachFileRequest{Name: "receipt"},
			want: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.req.contentType())
		})
	}
}

func TestGetTransactionAttachmentURL(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantURL    string
		wantErr    string
	}{
		{
			name:       "signed url",
			statusCode: http.StatusOK,
			response:   `{"url": "https://files.lunchmoney.app/66938-41ebb56a066bf09898de.png?X-Header1=X-Value1", "expires_at": "2025-07-14T12:00:00Z"}`,
			wantURL:    "https://files.lunchmoney.app/66938-41ebb56a066bf09898de.png?X-Header1=X-Value1",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   `{"message": "Not Found", "errors": [{"errMsg": "File attachment 1234567890 not found"}]}`,
			wantErr:    "File attachment 1234567890 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				// The file endpoints are not nested under the transaction.
				assert.Equal(t, "/transactions/attachments/1234567890", r.URL.Path)

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).GetTransactionAttachmentURL(context.Background(), 1234567890)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, got.URL)
			assert.Equal(t, time.Date(2025, 7, 14, 12, 0, 0, 0, time.UTC), got.ExpiresAt)
		})
	}
}

func TestDeleteTransactionAttachment(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    string
	}{
		{
			name:       "deleted",
			statusCode: http.StatusNoContent,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   `{"message": "Not Found", "errors": [{"errMsg": "File attachment 1234567890 not found"}]}`,
			wantErr:    "delete transaction attachment 1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "/transactions/attachments/1234567890", r.URL.Path)
				assert.Empty(t, r.URL.RawQuery)

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			err := testClient(t, server).DeleteTransactionAttachment(context.Background(), 1234567890)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
