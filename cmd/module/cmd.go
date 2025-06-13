package main

import (
	"go.viam.com/rdk/services/generic"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	"github.com/erh/viamstreamdeck"
)

func main() {

	arr := []resource.APIModel{
		resource.APIModel{generic.API, viamstreamdeck.ModelAny},
	}
	
	for _, m := range viamstreamdeck.Models {
		arr = append(arr, resource.APIModel{generic.API, m.Model})
	}
	
	module.ModularMain(arr...)
}
