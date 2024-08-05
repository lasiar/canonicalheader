package global

func dontImportPackage() {
	header.Get("Test-HEader") // want `use "Test-Header" instead of "Test-HEader"`
}
