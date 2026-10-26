package planner

import (
	"fmt"
	"math/bits"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// databaseCosts counts each physical PIR database, even when its column slices
// are represented implicitly. The same formulas serve evaluation and search.
func databaseCosts(c *pirconfig.Configuration, count uint64) (Costs, error) {
	var result Costs
	for _, term := range []struct {
		total *uint64
		a, b  uint64
		name  string
	}{
		{&result.StorageBits, c.Capacity.N, uint64(c.Code.S), "storage"},
		{&result.UploadBits, c.Query.Length, uint64(bits.Len64(c.Capacity.N - 1)), "upload"},
		{&result.DownloadBits, c.Query.Length, uint64(c.Code.S), "download"},
	} {
		value, err := multiply(term.a, term.b)
		if err == nil {
			value, err = multiply(value, count)
		}
		if err != nil {
			return Costs{}, fmt.Errorf("%s bits: %w", term.name, err)
		}
		*term.total = value
	}
	return finishCommunication(result)
}

// addModelCosts preserves a's application size and adds logical storage/traffic.
func addModelCosts(a, b Costs) (Costs, error) {
	for _, term := range []struct {
		total *uint64
		value uint64
		name  string
	}{
		{&a.StorageBits, b.StorageBits, "storage"}, {&a.UploadBits, b.UploadBits, "upload"}, {&a.DownloadBits, b.DownloadBits, "download"},
	} {
		if err := accumulate(term.total, term.value); err != nil {
			return Costs{}, fmt.Errorf("%s bits: %w", term.name, err)
		}
	}
	return finishCommunication(a)
}

func scaleModelCosts(c Costs, n uint64) (Costs, error) {
	var result Costs
	for _, term := range []struct {
		total *uint64
		value uint64
		name  string
	}{
		{&result.StorageBits, c.StorageBits, "storage"}, {&result.UploadBits, c.UploadBits, "upload"}, {&result.DownloadBits, c.DownloadBits, "download"},
	} {
		value, err := multiply(term.value, n)
		if err != nil {
			return Costs{}, fmt.Errorf("%s bits: %w", term.name, err)
		}
		*term.total = value
	}
	return finishCommunication(result)
}

func finishCommunication(c Costs) (Costs, error) {
	c.CommunicationBits = c.UploadBits
	if err := accumulate(&c.CommunicationBits, c.DownloadBits); err != nil {
		return Costs{}, fmt.Errorf("communication bits: %w", err)
	}
	return c, nil
}

func withinBudgets(c Costs, r Requirements) bool {
	return (r.MaxStorageBits == 0 || c.StorageBits <= r.MaxStorageBits) && (r.MaxUploadBits == 0 || c.UploadBits <= r.MaxUploadBits) && (r.MaxDownloadBits == 0 || c.DownloadBits <= r.MaxDownloadBits)
}

func cheaper(a, b Costs) bool {
	return a.StorageBits < b.StorageBits || (a.StorageBits == b.StorageBits && a.CommunicationBits < b.CommunicationBits)
}

// repeatTraffic scales indexed PIR queries while leaving stored data unchanged.
func repeatTraffic(c Costs, queries uint64) (Costs, error) {
	var err error
	c.UploadBits, err = multiply(c.UploadBits, queries)
	if err != nil {
		return Costs{}, fmt.Errorf("upload bits: %w", err)
	}
	c.DownloadBits, err = multiply(c.DownloadBits, queries)
	if err != nil {
		return Costs{}, fmt.Errorf("download bits: %w", err)
	}
	return finishCommunication(c)
}
