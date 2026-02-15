package exported

// UsedExported is called below — should not be flagged.
func UsedExported() int {
	return 1
}

// UnusedExported is never called.
func UnusedExported() string { // want `function "UnusedExported" is declared but never used in production code`
	return "dead"
}

// unusedUnexported is never called.
func unusedUnexported() {} // want `function "unusedUnexported" is declared but never used in production code`

// usedUnexported is called below — should not be flagged.
func usedUnexported() int {
	return 2
}

func init() {
	_ = UsedExported()
	_ = usedUnexported()
}
