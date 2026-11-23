package liftrs

import "testing"

func TestHostPhaseAdmission(t *testing.T) {
	c := commandFixture()
	p := hostArrayPhase("Newton", 8, 65536, 4*16384, 0, 0, 0)
	if p.Components["base_coefficients"] != 262144 {
		t.Fatal("coefficient slots")
	}
	polynomial, err := BinaryPolynomial(8)
	if err != nil {
		t.Fatal(err)
	}
	base, err := NewBinaryField(8, polynomial)
	if err != nil {
		t.Fatal(err)
	}
	tables := newBoundedPreparationProducts(base, 18)
	if p.Components["field_tables"] != 4*uint64(len(tables.log)+len(tables.powers)) {
		t.Fatal("duplicated power table inventory")
	}
	if err := c.admitHost(p); err != nil {
		t.Fatal(err)
	}
	c.RSSBytes = p.NamedBytes + c.HostReserveBytes
	if err := c.admitHost(p); err != nil {
		t.Fatal("exact allowance", err)
	}
	c.RSSBytes--
	if err := c.admitHost(p); err == nil {
		t.Fatal("one byte over budget")
	}
	c = commandFixture()
	c.HostReserveBytes = 0
	if err := c.Validate(); err == nil {
		t.Fatal("unspecified reserve")
	}
	roots, err := rootHostInventories(8, 0, []uint64{3442, 1507, 1541, 895}, 65536, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 4 || roots[2].Components["base_coefficients"] != 6*2048*4 {
		t.Fatal("rounded subgroup domain", roots)
	}
	for _, p := range roots {
		var total uint64
		for _, bytes := range p.Components {
			total += bytes
		}
		if total != p.NamedBytes {
			t.Fatal("inventory sum")
		}
	}
	// The source-derived s18 Newton envelope includes the product output as well
	// as P,G,G²;this is distinct from the historical input-only three-array count.
	wide := hostArrayPhase("Newton", 18, 1<<36, 4*(1<<32), 0, 0, 0)
	if wide.Components["base_coefficients"] != 64<<30 {
		t.Fatal("s18 units")
	}
}
