package lunchmoney

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// attachmentsPath is the collection the file_id endpoints hang off. It is not
// nested under the transaction the file belongs to.
const attachmentsPath = "/transactions/attachments"

// quoteEscaper escapes a filename for a Content-Disposition header the same way
// mime/multipart does for the parts it builds itself.
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// AttachFileRequest is a file to attach to a transaction. The API caps files at
// 10MB and accepts jpeg, png, pdf, heic and heif.
type AttachFileRequest struct {
	// Name is the filename the API stores and shows.
	Name string `validate:"required"`

	// File is read to the end and buffered in memory before the upload starts.
	File io.Reader `validate:"required"`

	// ContentType overrides the type derived from Name's extension, which is
	// worth setting for the formats Go has no mapping for.
	ContentType string

	// Notes is optional free text stored alongside the file.
	Notes string
}

// contentType is the caller's type, or the one Name's extension implies.
func (a *AttachFileRequest) contentType() string {
	if a.ContentType != "" {
		return a.ContentType
	}

	if t := mime.TypeByExtension(filepath.Ext(a.Name)); t != "" {
		return t
	}

	return "application/octet-stream"
}

// encode buffers the request as a multipart body and returns it with the
// content type naming its boundary.
func (a *AttachFileRequest) encode() (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, quoteEscaper.Replace(a.Name)))
	header.Set("Content-Type", a.contentType())

	part, err := w.CreatePart(header)
	if err != nil {
		return nil, "", fmt.Errorf("create file part: %w", err)
	}

	if _, err := io.Copy(part, a.File); err != nil {
		return nil, "", fmt.Errorf("read file: %w", err)
	}

	if a.Notes != "" {
		if err := w.WriteField("notes", a.Notes); err != nil {
			return nil, "", fmt.Errorf("write notes: %w", err)
		}
	}

	// Closing writes the trailing boundary, so the body is only complete once
	// this succeeds.
	if err := w.Close(); err != nil {
		return nil, "", fmt.Errorf("finish multipart body: %w", err)
	}

	return &buf, w.FormDataContentType(), nil
}

// TransactionAttachmentURL is a signed link to download an attachment, and the
// moment that link stops working.
type TransactionAttachmentURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AttachFileToTransaction uploads a file and attaches it to a transaction.
func (c *Client) AttachFileToTransaction(ctx context.Context, id int64, af *AttachFileRequest) (*TransactionAttachment, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, af); err != nil {
		return nil, err
	}

	buf, contentType, err := af.encode()
	if err != nil {
		return nil, fmt.Errorf("attach file to transaction %d: %w", id, err)
	}

	// Multipart cannot go through Post, which always marshals JSON.
	body, err := c.doBody(ctx, http.MethodPost, fmt.Sprintf("/transactions/%d/attachments", id), nil, buf, contentType)
	if err != nil {
		return nil, fmt.Errorf("attach file to transaction %d: %w", id, err)
	}

	resp := &TransactionAttachment{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// GetTransactionAttachmentURL returns a signed, expiring link to an attachment.
func (c *Client) GetTransactionAttachmentURL(ctx context.Context, fileID int64) (*TransactionAttachmentURL, error) {
	body, err := c.Get(ctx, fmt.Sprintf("%s/%d", attachmentsPath, fileID), nil)
	if err != nil {
		return nil, fmt.Errorf("get transaction attachment %d: %w", fileID, err)
	}

	resp := &TransactionAttachmentURL{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// DeleteTransactionAttachment detaches a file from its transaction and deletes it.
func (c *Client) DeleteTransactionAttachment(ctx context.Context, fileID int64) error {
	if _, err := c.Delete(ctx, fmt.Sprintf("%s/%d", attachmentsPath, fileID), nil); err != nil {
		return fmt.Errorf("delete transaction attachment %d: %w", fileID, err)
	}

	return nil
}
