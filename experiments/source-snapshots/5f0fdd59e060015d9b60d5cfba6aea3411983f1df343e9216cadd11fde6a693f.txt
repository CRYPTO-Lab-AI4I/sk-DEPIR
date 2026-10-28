//go:build liftrs_cpp && cgo

package liftrs

import liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"

func makeCPPPlan(transform *additiveTransform) (cppPlan, uint64, error) {
	words := cppDescription(transform)
	plan, err := liftrscpp.NewPlan(words)
	return plan, uint64(len(words)) * 4, err
}
