package pir

import (
	"fmt"
	legacy "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/rme_legacy/process"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/utils"
)

const rmEncodingIntermediate = legacy.RelativeIntermediate

// RMEncoder is a compatibility bridge for the old file-based PIR API. The RM
// implementation and process protocol live entirely under encoding/rme_legacy.
// It is not the standalone Lift-RS encoder in encoding/liftrs-shared.
type RMEncoder struct{ process legacy.RMEncoder }

// NewRMEncoder creates the legacy bridge without starting a process.
func NewRMEncoder() *RMEncoder { return &RMEncoder{} }

// Encode preserves the old RM-backed test path, applying the common PIR
// permutation and masks after legacy codeword generation. Keys must be initialized.
func (e *RMEncoder) Encode(input, output string, params Params, core *pirCore) (encoded Matrix, err error) {
	if core.permutator == nil || core.encryptor == nil {
		return nil, fmt.Errorf("PIR keys are not initialized; call Gen or GenFromConfig first")
	}
	err = e.process.Run(input, params.Q, params.K, params.M, func(path string) error {
		q := int(params.Q)
		switch params.M {
		case 2:
			data, err := utils.ReadMatrixFromFile(path)
			if err != nil {
				return err
			}
			encoded = &Matrix2D{data: data, q: q}
		case 3:
			data, err := utils.Read3DMatrixFromFile(path, q)
			if err != nil {
				return err
			}
			encoded = &Matrix3D{data: data, q: q}
		default:
			return fmt.Errorf("RMEncoding supports dimensions 2 and 3, got %d", params.M)
		}
		if err := core.permuteAndEncryptMatrix(0, encoded); err != nil {
			return err
		}
		switch matrix := encoded.(type) {
		case *Matrix2D:
			return utils.Write2DMatrixToFile(matrix.data, output)
		case *Matrix3D:
			return utils.Write3DMatrixToFile(matrix.data, output)
		}
		return fmt.Errorf("unsupported RMEncoding matrix type %T", encoded)
	})
	return
}

// Close releases the legacy subprocess, if one was started.
func (e *RMEncoder) Close() error { return e.process.Close() }
