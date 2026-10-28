// layout reports Go type sizes used by the static resource model. It allocates
// no encoder/preprocessing arrays and is not a memory or latency benchmark.
package main

import (
	"encoding/json"
	liftrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"os"
	"reflect"
)

func main() {
	tree, ok := reflect.TypeOf(liftrs.PaperEncoder{}).FieldByName("tree")
	if !ok {
		panic("product-tree field changed; update the resource inventory")
	}
	if err := json.NewEncoder(os.Stdout).Encode(map[string]uint64{
		"host_int_bytes":          uint64(reflect.TypeOf(int(0)).Size()),
		"product_tree_node_bytes": uint64(tree.Type.Elem().Size()),
	}); err != nil {
		panic(err)
	}
}
