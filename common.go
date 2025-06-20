package viamstreamdeck

import (
	"fmt"
	"image/color"
	"strings"

	"golang.org/x/image/colornames"

	"go.viam.com/rdk/resource"
)

func findDep(deps resource.Dependencies, n string) (resource.Resource, bool) {
	for nn, r := range deps {
		if nn.ShortName() == n {
			return r, true
		}
	}
	return nil, false
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	result := ""

	for _, part := range parts {
		result += strings.ToUpper(string(part[0])) + part[1:]
	}

	return result
}

func getColor(want, def string) color.Color {
	c, ok := colornames.Map[want]
	if ok {
		return c
	}

	c, ok = colornames.Map[def]
	if ok {
		return c
	}

	panic(fmt.Errorf("default color didn't work [%s]", def))
}
