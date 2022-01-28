package orchid

import "testing"

func TestINFO(t *testing.T) {
	Init("TestFramework")
	Info("INFO")
	OK("OK")
	Error("ERROR")
	Warn("WARNING")
	Debug("DEBUG")
}
