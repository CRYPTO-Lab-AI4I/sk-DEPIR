//go:build linux && liftrs_mpi && cgo

package liftrs

import (
	"encoding/binary"
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// MPI transport preserves descriptor and pre/post payload checks for every tile.
// It is shared by the internal GPU runtime and independent bounded test drivers.
type transposeMPI[Word transposeWord] struct {
	world  *mpi.World
	rank   int
	record func(string, int, transposePacket, []Word, [64]byte, [64]byte)
}

func transposeDescriptor[Word transposeWord](p transposePacket, source int) (record [64]byte) {
	var zero Word
	width := uint64(8)
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	for j, value := range []uint64{p.Operation, p.Rows, p.Columns, p.Round, p.Offset, p.Words, width, uint64(source)} {
		binary.LittleEndian.PutUint64(record[j*8:], value)
	}
	return
}
func (l transposeMPI[Word]) control(peer, source int, p transposePacket) error {
	expected := transposeDescriptor[Word](p, source)
	remote, err := l.world.ExchangeControl(peer, expected)
	if err != nil {
		return err
	}
	if remote != expected {
		return fmt.Errorf("transpose descriptor mismatch peer%d", peer)
	}
	return nil
}
func (l transposeMPI[Word]) SendTile(peer int, p transposePacket, tile []Word) error {
	if uint64(len(tile)) != p.Words {
		return fmt.Errorf("transpose send extent")
	}
	if err := l.control(peer, l.rank, p); err != nil {
		return err
	}
	before := mpi.PayloadFingerprint(tile)
	var err error
	switch values := any(tile).(type) {
	case []uint32:
		err = l.world.SendCoefficients32(peer, values)
	case []uint64:
		err = l.world.SendCoefficients64(peer, values)
	}
	if err != nil {
		return err
	}
	return l.world.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(tile), func(own, remote [64]byte) { l.record("send", peer, p, tile, own, remote) })
}
func (l transposeMPI[Word]) ReceiveTile(peer int, p transposePacket, tile []Word) error {
	if uint64(len(tile)) != p.Words {
		return fmt.Errorf("transpose receive extent")
	}
	if err := l.control(peer, peer, p); err != nil {
		return err
	}
	var err error
	switch values := any(tile).(type) {
	case []uint32:
		err = l.world.ReceiveCoefficients32(peer, values)
	case []uint64:
		err = l.world.ReceiveCoefficients64(peer, values)
	}
	if err != nil {
		return err
	}
	before := mpi.PayloadFingerprint(tile)
	return l.world.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(tile), func(own, remote [64]byte) { l.record("receive", peer, p, tile, own, remote) })
}
