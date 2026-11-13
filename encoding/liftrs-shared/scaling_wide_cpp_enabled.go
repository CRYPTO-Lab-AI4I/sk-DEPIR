//go:build liftrs_cpp && cgo

package liftrs

import cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"

type nativeWideCPP struct{ *cpp.WideOnline }

func makeWideCPPOnline(h ScalingCodeHeader) (wideCPPOnline, error) {
	p, err := cpp.NewWideOnline(h.Spec.FieldDegree, h.Polynomial, h.Omega, h.K)
	if err != nil {
		return nil, err
	}
	return nativeWideCPP{p}, nil
}
func (p nativeWideCPP) Resources() WideCPPResources {
	r := p.Layout()
	return WideCPPResources{r.Prefix, r.Arena, r.WorkspaceBytes, r.OutputBackingBytes, r.TransientBytes}
}
func (p nativeWideCPP) EncodeBorrowed(message, code, s32 []uint32, s64, arena []uint64, h []uint32, scratch []uint64, workers int) ([]uint32, WideCPPEncodingStats, error) {
	word, s, err := p.WideOnline.EncodeBorrowed(message, code, s32, s64, arena, h, scratch, workers)
	return word, WideCPPEncodingStats{Native: s.Native, Taylor: s.Taylor, Columns: s.Columns, Rows: s.Rows, Buffer: s.Buffer, Pointwise: s.Pointwise, Assembly: s.Assembly, Calls: s.Calls, OutputBackingBytes: s.OutputBackingBytes, TransientBytes: s.TransientBytes}, err
}
