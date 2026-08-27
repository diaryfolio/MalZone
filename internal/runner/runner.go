package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Result struct {
	Type                      string `json:"type"`
	Analysis                  string `json:"analysis"`
	SHA256                    string `json:"sha256"`
	Summary                   string `json:"summary"`
	ServiceAccountTokenAbsent bool   `json:"serviceAccountTokenAbsent"`
	KubernetesAPIDenied       bool   `json:"kubernetesApiDenied"`
}

func Run(ctx context.Context) error {
	analysis := os.Getenv("ANALYSIS_NAME")
	content := os.Getenv("CANARY_CONTENT")
	duration, err := strconv.Atoi(os.Getenv("RUN_SECONDS"))
	if err != nil || duration < 1 || duration > 60 {
		return fmt.Errorf("RUN_SECONDS must be between 1 and 60")
	}
	emit(map[string]any{"type": "runner.started", "analysis": analysis})
	// A label-selecting NetworkPolicy can take a short time to be programmed for a new Pod IP.
	// This harmless canary waits before measuring steady-state denial. This is not a production
	// pre-execution isolation mechanism and must never be used to justify running hostile content.
	if err := wait(ctx, 2*time.Second); err != nil {
		return err
	}
	tokenAbsent := fileAbsent(filepath.Join("/var/run/secrets/kubernetes.io/serviceaccount", "token"))
	apiDenied := kubernetesAPIDenied()
	if err := wait(ctx, time.Duration(duration)*time.Second); err != nil {
		return err
	}
	digest := sha256.Sum256([]byte(content))
	return emit(Result{
		Type:                      "result",
		Analysis:                  analysis,
		SHA256:                    hex.EncodeToString(digest[:]),
		Summary:                   "harmless lifecycle canary completed; no sample was executed",
		ServiceAccountTokenAbsent: tokenAbsent,
		KubernetesAPIDenied:       apiDenied,
	})
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func fileAbsent(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func kubernetesAPIDenied() bool {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	if port == "" {
		port = "443"
	}
	if host == "" {
		return true
	}
	connection, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 1200*time.Millisecond)
	if err != nil {
		return true
	}
	_ = connection.Close()
	return false
}

func emit(value any) error {
	return json.NewEncoder(os.Stdout).Encode(value)
}
