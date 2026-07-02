package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

// NoResponse tells the Respond function to not respond to the request. In these
// cases the app layer code has already done so.
type NoResponse struct{}

// NewNoResponse constructs a no response value.
func NewNoResponse() NoResponse {
	return NoResponse{}
}

// Encode implements the Encoder interface.
func (NoResponse) Encode() ([]byte, string, error) {
	return nil, "", nil
}

// =============================================================================

type httpStatus interface {
	HTTPStatus() int
}

// Respond sends a response to the client.
func Respond(ctx context.Context, w http.ResponseWriter, resp Encoder) error {
	if _, ok := resp.(NoResponse); ok {
		return nil
	}

	// If the context has been canceled, it means the client is no longer
	// waiting for a response.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	statusCode := http.StatusOK

	switch v := resp.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()

	case error:
		statusCode = http.StatusInternalServerError

	default:
		if resp == nil {
			statusCode = http.StatusNoContent
		}
	}

	_, span := addSpan(ctx, "web.send.response", attribute.Int("status", statusCode))
	defer span.End()

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := resp.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}

// ErrorResponse wraps an error and implements the Encoder interface
type ErrorResponse struct {
	Error   error  `json:"error"`
	Message string `json:"message,omitempty"`
	Status  int    `json:"status"`
}

func (e ErrorResponse) Encode() (data []byte, contentType string, err error) {
	data, err = json.Marshal(map[string]any{
		"error":   e.Error.Error(),
		"message": e.Message,
	})
	return data, "application/json", err
}

func (e ErrorResponse) HTTPStatus() int {
	return e.Status
}

// SuccessResponse wraps a success message and data
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (s SuccessResponse) Encode() (data []byte, contentType string, err error) {
	data, err = json.Marshal(s)
	return data, "application/json", err
}
