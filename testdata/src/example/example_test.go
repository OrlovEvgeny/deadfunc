package example

import "testing"

func TestOnlyUsedInTests(t *testing.T) {
	_ = onlyUsedInTests()
}
