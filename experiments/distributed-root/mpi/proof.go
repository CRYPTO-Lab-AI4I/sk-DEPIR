//go:build liftrs_mpi && cgo

package mpi

import (
	"bytes"
	"fmt"
)

// ProofMismatch checks local pre/post stability and cross-peer equality. These
// diagnostic fingerprints describe public experiment data, not authentication.
func ProofMismatch(own, remote [64]byte) bool {
	return !bytes.Equal(own[:32], own[32:]) || !bytes.Equal(remote[:32], remote[32:]) || !bytes.Equal(own[:32], remote[:32])
}

// CheckPayloadProof shares the tree and host-only diagnostic handshake. Capture
// must synchronously flush/close any mismatch files before returning. The second
// exchange then prevents an abort from killing a peer before its capture is saved.
func (w *World) CheckPayloadProof(peer int, before, after [32]byte, capture func([64]byte, [64]byte)) error {
	var own [64]byte
	copy(own[:32], before[:])
	copy(own[32:], after[:])
	remote, err := w.ExchangeProof(peer, own)
	if err != nil {
		return err
	}
	capture(own, remote)
	if ProofMismatch(own, remote) {
		if _, err = w.ExchangeProof(peer, own); err != nil {
			return err
		}
		return fmt.Errorf("transfer fingerprint mismatch peer%d", peer)
	}
	return nil
}
