// Package process isolates the historical RM subprocess protocol. It is not
// the Lift-RS encoder. Run serializes the fixed intermediate file through consume.
package process

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// RelativeIntermediate is the historical package-directory test path.
const RelativeIntermediate = "../output/inter.csv"

var rmEncodingProcessMu sync.Mutex

// RMEncoder owns one lazily started legacy encoder process.
type RMEncoder struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// NewRMEncoder constructs a lazy process; Close releases it after use.
func NewRMEncoder() *RMEncoder { return &RMEncoder{} }

// Run writes a raw RM codeword, then consumes it while holding the shared file
// lock. The callback must not retain the path for a later read.
func (e *RMEncoder) Run(input string, q uint64, k, m uint8, consume func(string) error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	rmEncodingProcessMu.Lock()
	defer rmEncodingProcessMu.Unlock()
	if e.cmd == nil {
		if err := e.start(); err != nil {
			return err
		}
	}
	path := RelativeIntermediate
	if info, err := os.Stat("output"); err == nil && info.IsDir() {
		path = filepath.Join("output", "inter.csv")
	}
	request := fmt.Sprintf("%s %s %d %d %d\n", input, path, q, k, m)
	if _, err := io.WriteString(e.stdin, request); err != nil {
		return fmt.Errorf("send RMEncoding request: %w", err)
	}
	var done [1]byte
	if _, err := io.ReadFull(e.stdout, done[:]); err != nil {
		return fmt.Errorf("wait for RMEncoding response: %w", err)
	}
	return consume(path)
}

// Close stops the lazily started RMEncoding process.
func (e *RMEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cmd == nil {
		return nil
	}
	if err := e.stdin.Close(); err != nil {
		return fmt.Errorf("close RMEncoding input: %w", err)
	}
	err := e.cmd.Wait()
	e.cmd = nil
	e.stdin = nil
	e.stdout = nil
	if err != nil {
		return fmt.Errorf("wait for RMEncoding process: %w", err)
	}
	return nil
}

func (e *RMEncoder) start() error {
	cmd, err := rmEncoderCommand()
	if err != nil {
		return err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("open RMEncoding stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open RMEncoding stdout: %w", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start RMEncoding process: %w", err)
	}
	e.cmd = cmd
	e.stdin = stdin
	e.stdout = stdout
	return nil
}

func rmEncoderCommand() (*exec.Cmd, error) {
	cppCandidates := []string{"encoding/rme_legacy/rme", "../encoding/rme_legacy/rme"}
	for _, cppBinary := range cppCandidates {
		if _, err := os.Stat(cppBinary); err == nil {
			return exec.Command(cppBinary), nil
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect C++ RMEncoding binary %q: %w", cppBinary, err)
		}
	}

	juliaPath, err := findJuliaPath()
	if err != nil {
		return nil, fmt.Errorf("find RMEncoding implementation: C++ binary is absent and Julia is unavailable: %w", err)
	}
	for _, script := range []string{"encoding/rme_legacy/GoRMEInterface.jl", "../encoding/rme_legacy/GoRMEInterface.jl"} {
		if _, err := os.Stat(script); err == nil {
			return exec.Command(juliaPath, "--project="+filepath.Dir(script), script), nil
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect Julia RMEncoding interface %q: %w", script, err)
		}
	}
	return nil, fmt.Errorf("find RMEncoding implementation: GoRMEInterface.jl is absent")
}

func findJuliaPath() (string, error) {
	cmd := exec.Command("which", "julia")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return "", err
	}
	path := strings.TrimSpace(output.String())
	if path == "" {
		return "", fmt.Errorf("which julia returned an empty path")
	}
	return path, nil
}
