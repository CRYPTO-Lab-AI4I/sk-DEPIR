//go:build linux && liftrs_mpi && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// Taylor headers have more fields than the fixed64-byte control frame. Hash all
// fields in canonical order, including source owner and slot width. This detects
// accidental phase/window mismatches; it is not authentication of public data.
func taylorDescriptor[Word transposeWord](p taylorPacket, source int) (out [64]byte) {
	var zero Word
	width := uint64(8)
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	q := uint64(0)
	if p.Quotient {
		q = 1
	}
	fields := [13]uint64{p.Operation, p.Domain, p.Columns, uint64(p.Level), q, p.Target.Block, p.Target.First, p.Target.Count, p.Slot, p.Source, p.Words, width, uint64(source)}
	var encoded [13 * 8]byte
	for j, v := range fields {
		binary.LittleEndian.PutUint64(encoded[j*8:], v)
	}
	digest := sha256.Sum256(encoded[:])
	copy(out[:32], digest[:])
	copy(out[32:], digest[:])
	return
}

type taylorMPI[Word transposeWord] struct {
	world  *mpi.World
	record func(string, int, taylorPacket, []Word, [64]byte, [64]byte)
}

func (l taylorMPI[Word]) Barrier() { l.world.Barrier() }
func (l taylorMPI[Word]) control(peer, source int, p taylorPacket) error {
	own := taylorDescriptor[Word](p, source)
	remote, err := l.world.ExchangeControl(peer, own)
	if err != nil {
		return err
	}
	if own != remote {
		return fmt.Errorf("Taylor window descriptor peer%d", peer)
	}
	return nil
}
func (l taylorMPI[Word]) SendWindow(peer int, p taylorPacket, data []Word) error {
	if uint64(len(data)) != p.Words {
		return fmt.Errorf("Taylor send extent")
	}
	if err := l.control(peer, l.world.Rank, p); err != nil {
		return err
	}
	before := mpi.PayloadFingerprint(data)
	var err error
	switch values := any(data).(type) {
	case []uint32:
		err = l.world.SendCoefficients32(peer, values)
	case []uint64:
		err = l.world.SendCoefficients64(peer, values)
	}
	if err != nil {
		return err
	}
	return l.world.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(data), func(own, remote [64]byte) { l.record("send", peer, p, data, own, remote) })
}
func (l taylorMPI[Word]) ReceiveWindow(peer int, p taylorPacket, data []Word) error {
	if uint64(len(data)) != p.Words {
		return fmt.Errorf("Taylor receive extent")
	}
	if err := l.control(peer, peer, p); err != nil {
		return err
	}
	var err error
	switch values := any(data).(type) {
	case []uint32:
		err = l.world.ReceiveCoefficients32(peer, values)
	case []uint64:
		err = l.world.ReceiveCoefficients64(peer, values)
	}
	if err != nil {
		return err
	}
	before := mpi.PayloadFingerprint(data)
	return l.world.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(data), func(own, remote [64]byte) { l.record("receive", peer, p, data, own, remote) })
}
