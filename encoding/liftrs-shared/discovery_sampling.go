package liftrs

func discoverySampleWidth(s int) int {
	if s > 8 {
		return 4
	}
	return 1
}
func discoverySampleLimit(s int) int { return 2 + (128+s-1)/s }

// Keep byte order/masking identical for ordinary and partitioned discovery.
func discoverySampleValue(data []byte, s int) uint32 {
	value := uint32(data[0])
	if s > 8 {
		value |= uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	}
	return value & ((uint32(1) << uint(s)) - 1)
}

// Both members of a Frobenius orbit choose the same owner without a global
// support list. This deterministic rule does not promise balanced rank loads.
func discoveryRootOwner(index, q, order uint64, ranks int) int {
	return int(min(index, index*q%order) % uint64(ranks))
}
