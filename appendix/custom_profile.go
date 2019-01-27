// +build ignore

package main

import (
	"fmt"
	"runtime/pprof"
)

func newProfIfNotDef(name string) *pprof.Profile {
	prof := pprof.Lookup(name)
	if prof == nil {
		prof = pprof.NewProfile(name)
	}
	return prof
}

func main() {
	prof := newProfIfNotDef("my_package_namespace")

	fmt.Println(prof.Name())
}
