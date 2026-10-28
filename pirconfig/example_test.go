package pirconfig_test

import (
	"fmt"
	"os"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

func ExampleLoad() {
	f, err := os.Open("configs_catalogue.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	c, err := pirconfig.Load(f)
	if err != nil {
		panic(err)
	}
	first := c.Configurations[0]
	fmt.Println(len(c.Configurations))
	fmt.Println(first.ID, first.EncodingStatus, first.Capacity.Status)
	// Output:
	// 13
	// binary-s08-m2-t8 supported verified
}
