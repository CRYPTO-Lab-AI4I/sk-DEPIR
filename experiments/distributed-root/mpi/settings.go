//go:build liftrs_mpi && cgo

package mpi

import "os"

// TransportSettings records only the explicit diagnostic allowlist, not the
// full environment. Requested values are not evidence of actual PML selection;
// native component logs provide that independent check.
func TransportSettings() map[string]string {
	result := make(map[string]string)
	for _, key := range []string{"OMPI_MCA_pml", "OMPI_MCA_btl", "OMPI_MCA_pml_base_verbose", "OMPI_MCA_btl_base_verbose", "UCX_LOG_LEVEL", "UCX_RCACHE_ENABLE"} {
		result[key] = os.Getenv(key)
	}
	return result
}
