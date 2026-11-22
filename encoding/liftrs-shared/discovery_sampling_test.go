package liftrs

import "testing"

func TestDiscoverySamplingAndOrbitOwnership(t *testing.T) {
	if discoverySampleWidth(8) != 1 || discoverySampleWidth(9) != 4 || discoverySampleValue([]byte{255}, 4) != 15 || discoverySampleValue([]byte{1, 1, 2, 3}, 9) != 257 {
		t.Fatal("sample width/order/mask")
	}
	for _, s := range []int{4, 6} {
		q := uint64(1) << uint(s)
		order := q*q - 1
		for _, ranks := range []int{1, 2, 3, 5} {
			for i := uint64(0); i < order; i++ {
				owner := discoveryRootOwner(i, q, order, ranks)
				if owner < 0 || owner >= ranks || owner != discoveryRootOwner(i*q%order, q, order, ranks) {
					t.Fatal("orbit ownership", i, ranks)
				}
			}
		}
	}
}
