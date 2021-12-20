package main

import (
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// doGet helper to GET requests
func doGet(t *testing.T, mux *chi.Mux, url string) *httptest.ResponseRecorder {
	return testutil.NewRequest().Get(url).WithAcceptJson().GoWithHTTPHandler(t, mux).Recorder
}

func TestGCPExecCreds(t *testing.T) {
	var err error

	r, err := getHandler()
	require.NoError(t, err)

	// this test requires that we have locally the envar
	t.Run("Test get creds", func(t *testing.T) {
		rr := doGet(t, r, "/creds")
		assert.Equal(t, http.StatusOK, rr.Code)

		response := make(map[string]interface{})
		err = json.NewDecoder(rr.Body).Decode(&response)
		assert.Equal(t, "client.authentication.k8s.io/v1beta1", response["apiVersion"])
		assert.Equal(t, "ExecCredential", response["kind"])
		assert.NotEmpty(t, response["status"])
		statusMap := response["status"].(map[string]interface{})
		assert.NotEmpty(t, statusMap["token"])
	})
}
