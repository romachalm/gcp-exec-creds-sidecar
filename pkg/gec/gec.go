package gec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
)

// Implement interface for openAPI
type GECServer struct{}

// ExecCredential is the Kubernetes ExecCredential API object
// Example:
// {
//   "apiVersion": "client.authentication.k8s.io/v1beta1",
//   "kind": "ExecCredential",
//   "status": {
//     "token": "my-bearer-token"
//   }
// }
type ExecCredential struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Status     ExecCredentialStatus `json:"status"`
}

type ExecCredentialStatus struct {
	Token string `json:"token"`
}

func NewExecCredential(token string) ExecCredential {
	ec := ExecCredential{
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Kind:       "ExecCredential",
		Status: ExecCredentialStatus{
			Token: token,
		},
	}
	return ec
}

func (g *GECServer) GetCreds(w http.ResponseWriter, r *http.Request) {

	logrus.Debug("got request...")

	clientScopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}

	tokenSource, err := google.DefaultTokenSource(context.Background(), clientScopes...)
	if err != nil {
		fmt.Printf("could not get tokenSource: %s", err)
		os.Exit(1)
	}

	token, err := tokenSource.Token()
	if err != nil {
		logrus.WithError(err).Error("failed to get token")
		os.Exit(1)
	}
	logrus.Debugf("token %s", token.AccessToken)

	ec := NewExecCredential(token.AccessToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ec); err != nil {
		logrus.WithError(err).Error("failed to encode JSON")
	}
}
