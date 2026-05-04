package runner

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lavianrose/flowforge/internal/config"
)

type DockerRunner struct {
	cli    *client.Client
	config config.DockerConfig
}

func NewDockerRunner(cfg config.DockerConfig) (*DockerRunner, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(cfg.Host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Verify Docker daemon is accessible
	_, err = cli.Ping(context.Background())
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("docker daemon not accessible: %w", err)
	}

	return &DockerRunner{cli: cli, config: cfg}, nil
}

func (r *DockerRunner) Close() error {
	if r.cli != nil {
		return r.cli.Close()
	}
	return nil
}

func (r *DockerRunner) Run(ctx context.Context, params RunParams) (*Result, error) {
	start := time.Now()

	img := r.resolveImage(params.Language)
	containerName := fmt.Sprintf("ff-%s-%s-%s-%d", trunc(params.TenantID, 8), trunc(params.RunID, 8), trunc(params.StepID, 8), start.UnixMilli())

	pidsLimit := int64(64)

	// Encode inputs as base64 to avoid shell escaping issues
	inputJSON, err := json.Marshal(params.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs: %w", err)
	}
	inputsB64 := base64.StdEncoding.EncodeToString(inputJSON)

	// Create container -- no stdin needed, all data via env vars
	createResp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image: img,
		Tty:   false,
		Env: []string{
			fmt.Sprintf("CODE=%s", params.Code),
			fmt.Sprintf("INPUTS_B64=%s", inputsB64),
		},
		Labels: map[string]string{
			"flowforge": "true",
			"tenant_id": params.TenantID,
			"run_id":    params.RunID,
			"step_id":   params.StepID,
		},
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:     r.config.DefaultMemoryMB * 1024 * 1024,
			MemorySwap: r.config.DefaultMemoryMB * 1024 * 1024,
			NanoCPUs:   int64(r.config.DefaultCPU * 1e9),
			PidsLimit:  &pidsLimit,
		},
		NetworkMode:    container.NetworkMode("none"),
		ReadonlyRootfs: true,
		SecurityOpt:    []string{"no-new-privileges:true"},
		CapDrop:        []string{"ALL"},
		Tmpfs: map[string]string{
			"/tmp": "rw,noexec,size=64m",
		},
		AutoRemove: false,
		LogConfig: container.LogConfig{
			Config: map[string]string{
				"max-size": "1m",
				"max-file": "1",
			},
		},
	}, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Guarantee cleanup
	defer func() {
		r.cli.ContainerRemove(context.Background(), createResp.ID, container.RemoveOptions{Force: true})
	}()

	// Start container
	if err := r.cli.ContainerStart(ctx, createResp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container with timeout
	timeout := time.Duration(r.config.DefaultTimeoutS) * time.Second
	waitCtx, waitCancel := context.WithTimeout(ctx, timeout)
	defer waitCancel()

	statusCh, errCh := r.cli.ContainerWait(waitCtx, createResp.ID, container.WaitConditionNotRunning)

	var exitCode int64

	select {
	case status := <-statusCh:
		exitCode = status.StatusCode
	case err := <-errCh:
		return nil, fmt.Errorf("container wait error: %w", err)
	case <-waitCtx.Done():
		// Timeout -- stop the container
		r.cli.ContainerStop(context.Background(), createResp.ID, container.StopOptions{})
		return &Result{
			Stderr:   fmt.Sprintf("execution timed out after %s", timeout),
			TimedOut: true,
			Duration: time.Since(start),
			ExitCode: -1,
		}, fmt.Errorf("script execution timed out after %s", timeout)
	}

	// Inspect container to check OOM status
	inspectResp, err := r.cli.ContainerInspect(context.Background(), createResp.ID)
	oomKilled := err == nil && inspectResp.State != nil && inspectResp.State.OOMKilled

	// Get output via logs API
	logsReader, err := r.cli.ContainerLogs(context.Background(), createResp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logsReader.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	stdcopy.StdCopy(&stdoutBuf, &stderrBuf, logsReader)

	result := &Result{
		ExitCode:  int(exitCode),
		OOMKilled: oomKilled,
		Stderr:    strings.TrimSpace(stderrBuf.String()),
		Duration:  time.Since(start),
	}

	if oomKilled {
		return result, fmt.Errorf("script exceeded memory limit (%dMB)", r.config.DefaultMemoryMB)
	}

	if exitCode != 0 {
		stderr := result.Stderr
		if stderr == "" {
			stderr = "(no stderr output)"
		}
		return result, fmt.Errorf("script exited with code %d: %s", exitCode, stderr)
	}

	// Parse stdout as JSON
	stdout := strings.TrimSpace(stdoutBuf.String())
	if stdout == "" {
		return nil, fmt.Errorf("script produced no output")
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		// stdout is not a JSON object -- wrap raw output in a result map
		result.Output = map[string]interface{}{"output": stdout}
		return result, nil
	}

	result.Output = output
	return result, nil
}

// CleanupOrphaned removes any containers labeled with flowforge=true
// that may have been left from a previous crash.
func (r *DockerRunner) CleanupOrphaned(ctx context.Context) error {
	f := filters.NewArgs()
	f.Add("label", "flowforge=true")

	containers, err := r.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		r.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
	}

	return nil
}

// PullImages pulls the runner images if not present locally.
func (r *DockerRunner) PullImages(ctx context.Context) error {
	for _, img := range []string{r.config.PythonImage, r.config.NodeImage} {
		reader, err := r.cli.ImagePull(ctx, img, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image %s: %w", img, err)
		}
		io.Copy(io.Discard, reader)
		reader.Close()
	}
	return nil
}

func (r *DockerRunner) resolveImage(language string) string {
	switch language {
	case "python":
		return r.config.PythonImage
	case "javascript":
		return r.config.NodeImage
	default:
		return r.config.PythonImage
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
