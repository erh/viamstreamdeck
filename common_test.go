package viamstreamdeck

import (
	"testing"

	"go.viam.com/test"
)

func TestSnakeToCamel(t *testing.T) {
	test.That(t, snakeToCamel("foo_bar"), test.ShouldEqual, "FooBar")
}
