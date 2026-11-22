//go:build linux && liftrs_mpi && cgo

package liftrs

import (
	"encoding/binary"
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

type operandMPI = typedOperandMPI[uint32]

type typedOperandMPI[Word transposeWord] struct {
	world  *mpi.World
	record func(string, operandPacket, []Word, [64]byte, [64]byte)
}

// Bind each payload to its polynomial, global coefficient interval and owners.
// Existing MPI coefficient/proof helpers preserve the established wire codec.
func operandDescriptor[Word transposeWord](p operandPacket) (d [64]byte) {
	width := uint64(4)
	var zero Word
	if _, ok := any(zero).(uint64); ok {
		width = 8
	}
	for j, v := range []uint64{p.Operation, p.Domain, p.Operand, p.Offset, p.Words, uint64(p.Source), uint64(p.Target), width} {
		binary.LittleEndian.PutUint64(d[j*8:], v)
	}
	return
}
func (l typedOperandMPI[Word]) transfer(p operandPacket, values []Word, send bool) error {
	owner, peer := p.Target, p.Source
	if send {
		owner, peer = p.Source, p.Target
	}
	if l.world == nil || l.record == nil || l.world.Rank != owner || p.Words == 0 || p.Words != uint64(len(values)) || p.Source == p.Target || p.Source < 0 || p.Target < 0 || p.Source >= l.world.Size || p.Target >= l.world.Size || p.Operation == 0 || p.Operand > 1 || p.Offset > p.Domain || p.Words > p.Domain-p.Offset {
		return fmt.Errorf("operand MPI packet/ownership")
	}
	expected := operandDescriptor[Word](p)
	remote, err := l.world.ExchangeControl(peer, expected)
	if err != nil {
		return err
	}
	if remote != expected {
		return fmt.Errorf("operand MPI descriptor mismatch")
	}
	before := mpi.PayloadFingerprint(values)
	direction := "send"
	if send {
		switch v := any(values).(type) {
		case []uint32:
			err = l.world.SendCoefficients32(peer, v)
		case []uint64:
			err = l.world.SendCoefficients64(peer, v)
		}
	} else {
		direction = "receive"
		switch v := any(values).(type) {
		case []uint32:
			err = l.world.ReceiveCoefficients32(peer, v)
		case []uint64:
			err = l.world.ReceiveCoefficients64(peer, v)
		}
		before = mpi.PayloadFingerprint(values)
	}
	if err != nil {
		return err
	}
	return l.world.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(values), func(own, remote [64]byte) { l.record(direction, p, values, own, remote) })
}
func (l typedOperandMPI[Word]) SendOperand(p operandPacket, v []Word) error {
	return l.transfer(p, v, true)
}
func (l typedOperandMPI[Word]) ReceiveOperand(p operandPacket, v []Word) error {
	return l.transfer(p, v, false)
}
