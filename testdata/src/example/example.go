package example

// usedFunc is called in production code below.
func usedFunc() int {
	return 42
}

// unusedFunc is never called anywhere in production.
func unusedFunc() string { // want `function "unusedFunc" is declared but never used in production code`
	return "dead"
}

// onlyUsedInTests is called only in _test.go files.
func onlyUsedInTests() bool { // want `function "onlyUsedInTests" is declared but never used in production code`
	return true
}

// init is always skipped.
func init() {
	_ = usedFunc()
	useThings()
}

// helperCalledByHelper is called by another helper.
func helperCalledByHelper() int {
	return 1
}

// callerHelper calls helperCalledByHelper.
func callerHelper() int {
	return helperCalledByHelper()
}

// unusedHelper is not called by anyone.
func unusedHelper() {} // want `function "unusedHelper" is declared but never used in production code`

// ExportedFunc is exported — skipped by default.
func ExportedFunc() {}

// --- Methods ---

type myStruct struct{}

// usedMethod is called in production code.
func (m *myStruct) usedMethod() int { return 1 }

// unusedMethod is never called.
func (m *myStruct) unusedMethod() int { return 2 } // want `function "myStruct.unusedMethod" is declared but never used in production code`

// String implements fmt.Stringer — should be skipped.
func (m *myStruct) String() string { return "my" }

// Error implements the error interface — should be skipped.
func (m *myStruct) Error() string { return "err" }

func useThings() {
	_ = usedFunc()
	_ = callerHelper()

	s := &myStruct{}
	_ = s.usedMethod()
}
