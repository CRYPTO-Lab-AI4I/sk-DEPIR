//go:build linux && liftrs_mpi && cgo

package liftrs

import (
	"encoding/binary"
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

type operandMPI struct {
	world  *mpi.World
	record func(string, operandPacket, []uint32, [64]byte, [64]byte)
}

// Bind each payload to its polynomial, global coefficient interval and owners.
// Existing MPI coefficient/proof helpers preserve the established wire codec.
func operandDescriptor(p operandPacket) (d [64]byte) {
	for j, v := range []uint64{p.Operation, p.Domain, p.Operand, p.Offset, p.Words, uint64(p.Source), uint64(p.Target), 4} {
		binary.LittleEndian.PutUint64(d[j*8:], v)
	}
	return
}
func (l operandMPI) transfer(p operandPacket, values []uint32, send bool) error {
	owner, peer := p.Target, p.Source
	if send {
		owner, peer = p.Source, p.Target
	}
	if l.world == nil || l.record == nil || l.world.Rank != owner || p.Words == 0 || p.Words != uint64(len(values)) || p.Source == p.Target || p.Source < 0 || p.Target < 0 || p.Source >= l.world.Size || p.Target >= l.world.Size || p.Operation == 0 || p.Operand > 1 || p.Offset > p.Domain || p.Words > p.Domain-p.Offset {
		return fmt.Errorf("operand MPI packet/ownership")
	}
	expected := operandDescriptor(p)
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
		err = l.world.SendCoefficients32(peer, values)
	} else {
		direction = "receive"
		err = l.world.ReceiveCoefficients32(peer, values)
		before = mpi.PayloadFingerprint(values)
	}
	if err != nil {
		return err
	}
	return l.world.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(values), func(own, remote [64]byte) { l.record(direction, p, values, own, remote) })
}
func (l operandMPI) SendOperand(p operandPacket, v []uint32) error    { return l.transfer(p, v, true) }
func (l operandMPI) ReceiveOperand(p operandPacket, v []uint32) error { return l.transfer(p, v, false) }
