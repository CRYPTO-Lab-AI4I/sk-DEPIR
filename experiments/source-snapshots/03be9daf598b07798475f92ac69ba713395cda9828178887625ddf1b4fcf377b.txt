package liftrs

// exponentSet stores public code-membership bits. Callers validate indices before
// use; one bit per field position avoids hash-map or byte-per-position overhead.
type exponentSet []uint64

func newExponentSet(positions int) exponentSet { return make(exponentSet, (positions+63)/64) }
func (s exponentSet) has(i int) bool           { return s[i>>6]&(uint64(1)<<uint(i&63)) != 0 }
func (s exponentSet) add(i int)                { s[i>>6] |= uint64(1) << uint(i&63) }
