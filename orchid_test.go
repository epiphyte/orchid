// Package orchid
// Copyright (c) 2022 Epiphyte LLC. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Author: Fernandez-Alcon, Jose
// e-mail: jose@epiphyte.io
package orchid

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestINFO(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	var logger Logger

	logger.Init("TestFramework", "")
	logger.Info("Test message")
	res := strings.Split(buf.String(), " ")
	if len(res) < 5 || res[4] != "TestFramework" {
		t.Errorf("Expected TestFramework in position 4, got: %v", res)
	}
	//for i, s := range res {
	//	fmt.Println(i, s)
	//}
	//OK("OK")
	//Error("ERROR")
	//Warn("WARNING")
	//Debug("DEBUG")
}
