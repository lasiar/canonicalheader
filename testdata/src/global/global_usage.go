package global

func dontImportPackage() {
	header.Get("Test-HEader") // want `use canonical header key "Test-Header" instead of "Test-HEader"`
}
