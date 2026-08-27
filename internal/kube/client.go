package kube

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/diaryfolio/malzone/internal/model"
)

const serviceAccountRoot = "/var/run/secrets/kubernetes.io/serviceaccount"

type Error struct {
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("kubernetes API returned %d: %s", e.StatusCode, e.Body)
}

func IsNotFound(err error) bool {
	var apiError *Error
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound
}

func IsConflict(err error) bool {
	var apiError *Error
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusConflict
}

type Client struct {
	baseURL   string
	namespace string
	token     string
	http      *http.Client
}

func NewInCluster(namespace string) (*Client, error) {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	if port == "" {
		port = "443"
	}
	if host == "" {
		return nil, errors.New("KUBERNETES_SERVICE_HOST is not set")
	}
	token, err := os.ReadFile(filepath.Join(serviceAccountRoot, "token"))
	if err != nil {
		return nil, fmt.Errorf("read service-account token: %w", err)
	}
	caData, err := os.ReadFile(filepath.Join(serviceAccountRoot, "ca.crt"))
	if err != nil {
		return nil, fmt.Errorf("read cluster CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caData) {
		return nil, errors.New("cluster CA is invalid")
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
	}
	return &Client{
		baseURL:   "https://" + host + ":" + port,
		namespace: namespace,
		token:     strings.TrimSpace(string(token)),
		http:      &http.Client{Transport: transport, Timeout: 10 * time.Second},
	}, nil
}

func NewForTest(baseURL, namespace string, client *http.Client) *Client {
	return &Client{baseURL: baseURL, namespace: namespace, http: client}
}

func (c *Client) analysesPath() string {
	return "/apis/malzone.io/v1alpha1/namespaces/" + url.PathEscape(c.namespace) + "/analyses"
}

func (c *Client) ListAnalyses(ctx context.Context) ([]model.Analysis, error) {
	var list model.AnalysisList
	if err := c.do(ctx, http.MethodGet, c.analysesPath(), nil, "", &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) GetAnalysis(ctx context.Context, name string) (model.Analysis, error) {
	var analysis model.Analysis
	err := c.do(ctx, http.MethodGet, c.analysesPath()+"/"+url.PathEscape(name), nil, "", &analysis)
	return analysis, err
}

func (c *Client) CreateAnalysis(ctx context.Context, analysis model.Analysis) (model.Analysis, error) {
	var created model.Analysis
	err := c.do(ctx, http.MethodPost, c.analysesPath(), analysis, "application/json", &created)
	return created, err
}

func (c *Client) RequestCancel(ctx context.Context, name string) (model.Analysis, error) {
	patch := map[string]any{"spec": map[string]any{"cancelRequested": true}}
	var analysis model.Analysis
	err := c.do(ctx, http.MethodPatch, c.analysesPath()+"/"+url.PathEscape(name), patch,
		"application/merge-patch+json", &analysis)
	return analysis, err
}

func (c *Client) UpdateStatus(ctx context.Context, analysis model.Analysis) (model.Analysis, error) {
	body := map[string]any{
		"apiVersion": model.APIVersion,
		"kind":       model.Kind,
		"metadata": map[string]any{
			"name":            analysis.Metadata.Name,
			"resourceVersion": analysis.Metadata.ResourceVersion,
		},
		"status": analysis.Status,
	}
	var updated model.Analysis
	err := c.do(ctx, http.MethodPut, c.analysesPath()+"/"+url.PathEscape(analysis.Metadata.Name)+"/status",
		body, "application/json", &updated)
	return updated, err
}

type JobState struct {
	Active    int
	Succeeded int
	Failed    int
}

func (c *Client) GetJob(ctx context.Context, name string) (JobState, error) {
	var response struct {
		Status struct {
			Active    int `json:"active"`
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
		} `json:"status"`
	}
	path := "/apis/batch/v1/namespaces/" + url.PathEscape(c.namespace) + "/jobs/" + url.PathEscape(name)
	if err := c.do(ctx, http.MethodGet, path, nil, "", &response); err != nil {
		return JobState{}, err
	}
	return JobState(response.Status), nil
}

func (c *Client) CreateJob(ctx context.Context, analysis model.Analysis, image string) error {
	name := JobName(analysis.Metadata.Name)
	labels := map[string]string{
		"app.kubernetes.io/name":        "malzone",
		"app.kubernetes.io/component":   "poc-runner",
		"malzone.io/analysis":           analysis.Metadata.Name,
		"malzone.io/execution-boundary": "poc-only",
	}
	metadata := map[string]any{
		"name":      name,
		"namespace": c.namespace,
		"labels":    labels,
	}
	if analysis.Metadata.UID != "" {
		metadata["ownerReferences"] = []any{map[string]any{
			"apiVersion":         model.APIVersion,
			"kind":               model.Kind,
			"name":               analysis.Metadata.Name,
			"uid":                analysis.Metadata.UID,
			"controller":         true,
			"blockOwnerDeletion": true,
		}}
	}
	job := map[string]any{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata":   metadata,
		"spec": map[string]any{
			"backoffLimit":          0,
			"activeDeadlineSeconds": analysis.Spec.TimeoutSeconds + 15,
			"template": map[string]any{
				"metadata": map[string]any{"labels": labels},
				"spec": map[string]any{
					"automountServiceAccountToken":  false,
					"restartPolicy":                 "Never",
					"terminationGracePeriodSeconds": 5,
					"serviceAccountName":            "malzone-runner",
					"securityContext": map[string]any{
						"runAsNonRoot": true,
						"runAsUser":    65532,
						"runAsGroup":   65532,
						"seccompProfile": map[string]any{
							"type": "RuntimeDefault",
						},
					},
					"containers": []any{map[string]any{
						"name":    "runner",
						"image":   image,
						"command": []string{"/malzone", "runner"},
						"env": []any{
							map[string]any{"name": "ANALYSIS_NAME", "value": analysis.Metadata.Name},
							map[string]any{"name": "CANARY_CONTENT", "value": analysis.Spec.Sample.Content},
							map[string]any{"name": "RUN_SECONDS", "value": fmt.Sprintf("%d", analysis.Spec.TimeoutSeconds)},
						},
						"securityContext": map[string]any{
							"allowPrivilegeEscalation": false,
							"readOnlyRootFilesystem":   true,
							"capabilities":             map[string]any{"drop": []string{"ALL"}},
						},
						"resources": map[string]any{
							"requests": map[string]string{"cpu": "10m", "memory": "16Mi"},
							"limits":   map[string]string{"cpu": "100m", "memory": "32Mi"},
						},
						"volumeMounts": []any{map[string]any{"name": "scratch", "mountPath": "/tmp"}},
					}},
					"volumes": []any{map[string]any{"name": "scratch", "emptyDir": map[string]any{"sizeLimit": "1Mi"}}},
				},
			},
		},
	}
	path := "/apis/batch/v1/namespaces/" + url.PathEscape(c.namespace) + "/jobs"
	return c.do(ctx, http.MethodPost, path, job, "application/json", nil)
}

func (c *Client) DeleteJob(ctx context.Context, name string) error {
	path := "/apis/batch/v1/namespaces/" + url.PathEscape(c.namespace) + "/jobs/" + url.PathEscape(name)
	options := map[string]any{"apiVersion": "v1", "kind": "DeleteOptions", "propagationPolicy": "Foreground"}
	err := c.do(ctx, http.MethodDelete, path, options, "application/json", nil)
	if IsNotFound(err) {
		return nil
	}
	return err
}

func (c *Client) JobLogs(ctx context.Context, jobName string) (string, error) {
	var pods struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	query := url.Values{"labelSelector": []string{"job-name=" + jobName}, "limit": []string{"1"}}
	path := "/api/v1/namespaces/" + url.PathEscape(c.namespace) + "/pods?" + query.Encode()
	if err := c.do(ctx, http.MethodGet, path, nil, "", &pods); err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", errors.New("runner pod not found")
	}
	logPath := "/api/v1/namespaces/" + url.PathEscape(c.namespace) + "/pods/" +
		url.PathEscape(pods.Items[0].Metadata.Name) + "/log?container=runner&limitBytes=16384"
	var raw string
	if err := c.do(ctx, http.MethodGet, logPath, nil, "", &raw); err != nil {
		return "", err
	}
	return raw, nil
}

func JobName(analysisName string) string {
	const prefix = "mz-poc-"
	max := 63 - len(prefix)
	if len(analysisName) > max {
		analysisName = analysisName[:max]
	}
	return prefix + strings.TrimRight(analysisName, "-")
}

func (c *Client) do(ctx context.Context, method, path string, body any, contentType string, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &Error{StatusCode: response.StatusCode, Body: strings.TrimSpace(string(responseBody))}
	}
	if out == nil || len(responseBody) == 0 {
		return nil
	}
	if raw, ok := out.(*string); ok {
		*raw = string(responseBody)
		return nil
	}
	return json.Unmarshal(responseBody, out)
}
