package pir

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"rme/utils"
)

const rmEncodingIntermediate = "../output/inter.csv"

// The external encoder protocol writes through a fixed intermediate file.
// Serialize all RMEncoder instances so concurrent encodes cannot overwrite it.
var rmEncodingProcessMu sync.Mutex

// RMEncoder owns the external RMEncoding process used as a Lift-RS testing placeholder.
type RMEncoder struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// NewRMEncoder creates an RMEncoding process wrapper. The process starts lazily
// on the first Encode call.
func NewRMEncoder() *RMEncoder {
	return &RMEncoder{}
}

// Encode applies RMEncoding, then the common secret permutation and masking.
func (e *RMEncoder) Encode(input, output string, params Params, core *pirCore) (Matrix, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	rmEncodingProcessMu.Lock()
	defer rmEncodingProcessMu.Unlock()

	if core.permutator == nil || core.encryptor == nil {
		return nil, fmt.Errorf("PIR keys are not initialized; call Gen or GenFromConfig first")
	}
	if e.cmd == nil {
		if err := e.start(); err != nil {
			return nil, err
		}
	}

	encoded, err := e.encodeRMCodeword(input, rmEncodingIntermediatePath(), params)
	if err != nil {
		return nil, err
	}
	// TODO: Support sliced encoding and pass the slice ID to this operation.
	if err := core.permuteAndEncryptMatrix(0, encoded); err != nil {
		return nil, err
	}

	switch matrix := encoded.(type) {
	case *Matrix2D:
		if err := utils.Write2DMatrixToFile(matrix.data, output); err != nil {
			return nil, fmt.Errorf("write 2D RMEncoding output: %w", err)
		}
	case *Matrix3D:
		if err := utils.Write3DMatrixToFile(matrix.data, output); err != nil {
			return nil, fmt.Errorf("write 3D RMEncoding output: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported RMEncoding matrix type %T", encoded)
	}
	return encoded, nil
}

func rmEncodingIntermediatePath() string {
	if info, err := os.Stat("output"); err == nil && info.IsDir() {
		return filepath.Join("output", "inter.csv")
	}
	return rmEncodingIntermediate
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

func (e *RMEncoder) encodeRMCodeword(input, output string, params Params) (Matrix, error) {
	request := fmt.Sprintf("%s %s %d %d %d\n", input, output, params.Q, params.K, params.M)
	if _, err := io.WriteString(e.stdin, request); err != nil {
		return nil, fmt.Errorf("send RMEncoding request: %w", err)
	}
	var done [1]byte
	if _, err := io.ReadFull(e.stdout, done[:]); err != nil {
		return nil, fmt.Errorf("wait for RMEncoding response: %w", err)
	}

	q := int(params.Q)
	switch params.M {
	case 2:
		matrix, err := utils.ReadMatrixFromFile(output)
		if err != nil {
			return nil, fmt.Errorf("read 2D RMEncoding output: %w", err)
		}
		return &Matrix2D{data: matrix, q: q}, nil
	case 3:
		matrix, err := utils.Read3DMatrixFromFile(output, q)
		if err != nil {
			return nil, fmt.Errorf("read 3D RMEncoding output: %w", err)
		}
		return &Matrix3D{data: matrix, q: q}, nil
	default:
		return nil, fmt.Errorf("RMEncoding supports dimensions 2 and 3, got %d", params.M)
	}
}

func rmEncoderCommand() (*exec.Cmd, error) {
	cppCandidates := []string{"encoding/rme", "../encoding/rme"}
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
	for _, script := range []string{"encoding/GoRMEInterface.jl", "../encoding/GoRMEInterface.jl"} {
		if _, err := os.Stat(script); err == nil {
			return exec.Command(juliaPath, script), nil
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
