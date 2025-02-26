package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRespond(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		mockMarshal    func(v interface{}) ([]byte, error)
		mockWrite      func(w http.ResponseWriter, response []byte) (int, error)
		expectedBody   string
	}{

		{
			name:           "JSON Marshal error",
			payload:        struct{ Name chan int }{}, // Invalid type for JSON
			expectedStatus: http.StatusInternalServerError,
			mockMarshal:    func(v interface{}) ([]byte, error) { return nil, errors.New("marshal error") },
			mockWrite:      func(w http.ResponseWriter, response []byte) (int, error) { return w.Write(response) },
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name:           "Write response error",
			payload:        map[string]string{"message": "ok"},
			expectedStatus: http.StatusOK, // ✅ Expect 200, since status can't change after headers sent
			mockMarshal:    json.Marshal,
			mockWrite:      func(w http.ResponseWriter, response []byte) (int, error) { return 0, errors.New("forced write error") },
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			marshalJson = tc.mockMarshal
			writeResponse = tc.mockWrite
			defer func() {
				marshalJson = json.Marshal
				writeResponse = func(w http.ResponseWriter, response []byte) (int, error) { return w.Write(response) }
			}()
			recorder := httptest.NewRecorder()
			Respond(recorder, http.StatusOK, tc.payload)
			result := recorder.Result()
			defer result.Body.Close()
			require.Equal(t, tc.expectedStatus, result.StatusCode)
			body := readResponseBody(t, result)
			require.JSONEq(t, tc.expectedBody, body)
		})
	}
}

func TestRespondAfterFlush(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		mockMarshal    func(v interface{}) ([]byte, error)
		mockWrite      func(w http.ResponseWriter, response []byte) (int, error)
		expectedBody   string
	}{
		{
			name:           "JSON Marshal error",
			payload:        struct{ Name chan int }{}, // Invalid type for JSON
			expectedStatus: http.StatusInternalServerError,
			mockMarshal:    func(v interface{}) ([]byte, error) { return nil, errors.New("marshal error") },
			mockWrite:      func(w http.ResponseWriter, response []byte) (int, error) { return w.Write(response) },
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name:           "Write response error",
			payload:        map[string]string{"message": "ok"},
			expectedStatus: http.StatusInternalServerError,
			mockMarshal:    json.Marshal,
			mockWrite:      func(w http.ResponseWriter, response []byte) (int, error) { return 0, errors.New("forced write error") },
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mock marshalJson and writeResponse
			marshalJson = tc.mockMarshal
			writeResponse = tc.mockWrite
			defer func() {
				marshalJson = json.Marshal
				writeResponse = func(w http.ResponseWriter, response []byte) (int, error) { return w.Write(response) }
			}()

			// Set up recorder
			recorder := httptest.NewRecorder()
			RespondAfterFlush(recorder, tc.payload)

			result := recorder.Result()
			defer result.Body.Close()

			require.Equal(t, tc.expectedStatus, result.StatusCode)

			body := readResponseBody(t, result)
			require.JSONEq(t, tc.expectedBody, body)
		})
	}
}

// readResponseBody reads and returns the response body as a string.
func readResponseBody(t *testing.T, result *http.Response) string {
	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	return string(body)
}
