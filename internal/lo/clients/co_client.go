package lo

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type HTTPReporter struct {
    endpoint string        // base URL of CO API
    client   *http.Client
    timeout  time.Duration
}

func NewHTTPReporter(endpoint string, timeout time.Duration) *HTTPReporter {
    return &HTTPReporter{
        endpoint: endpoint,
        timeout:  timeout,
        client:   &http.Client{Timeout: timeout},
    }
}

func (r *HTTPReporter) ReportState(deploymentID string, state DeploymentState) error {
    ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
    defer cancel()

    body, err := json.Marshal(state)
    if err != nil {
        return fmt.Errorf("failed to marshal DeploymentState: %w", err)
    }

    url := fmt.Sprintf("%s/deployments/%s/state", r.endpoint, deploymentID)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := r.client.Do(req)
    if err != nil {
        return fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("failed to report state, status: %s", resp.Status)
    }

    fmt.Printf("[HTTPReporter] Successfully reported state for deployment %s\n", deploymentID)
    return nil
}
