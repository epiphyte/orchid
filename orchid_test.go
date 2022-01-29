// Package orchid
//Copyright (c) 2022 Epiphyte LLC. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Author: Fernandez-Alcon, Jose
// e-mail: jose@epiphyte.io
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
