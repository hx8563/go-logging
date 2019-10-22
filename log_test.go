// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"bytes"
	//	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestLogCalldepth(t *testing.T) {
	buf := &bytes.Buffer{}
	SetBackend(NewLogBackend(buf, "", log.Lshortfile))
	SetFormatter(MustStringFormatter("%{shortfile} %{level} %{message}"))

	log := MustGetLogger("test")
	log.Info("test filename")

	parts := strings.SplitN(buf.String(), " ", 2)

	// Verify that the correct filename is registered by the stdlib logger
	if !strings.HasPrefix(parts[0], "log_test.go:") {
		t.Errorf("incorrect filename: %s", parts[0])
	}
	// Verify that the correct filename is registered by go-logging
	if !strings.HasPrefix(parts[1], "log_test.go:") {
		t.Errorf("incorrect filename: %s", parts[1])
	}
}

func c(log *Logger) { log.Info("test callpath") }
func b(log *Logger) { c(log) }
func a(log *Logger) { b(log) }

func rec(log *Logger, r int) {
	if r == 0 {
		a(log)
		return
	}
	rec(log, r-1)
}

func testCallpath(t *testing.T, testId int, format string, expect string) {
	buf := &bytes.Buffer{}
	SetBackend(NewLogBackend(buf, "", log.Lshortfile))
	SetFormatter(MustStringFormatter(format))

	logger := MustGetLogger("test")
	rec(logger, 6)

	parts := strings.SplitN(buf.String(), " ", 3)
	//fmt.Println(buf)

	// Verify that the correct filename is registered by the stdlib logger
	if !strings.HasPrefix(parts[0], "log_test.go:") {
		t.Errorf("incorrect filename: %d %s %s", testId, parts[0], expect)
	}
	// Verify that the correct callpath is registered by go-logging
	if !strings.HasPrefix(parts[1], expect) {
		t.Errorf("incorrect callpath: %d %s %s", testId, parts[1], expect)
	}
	// Verify that the correct message is registered by go-logging
	if !strings.HasPrefix(parts[2], "test callpath") {
		t.Errorf("incorrect message: %d %s %s", testId, parts[2], expect)
	}
}

func TestLogCallpath(t *testing.T) {
	testCallpath(t, 0, "%{callpath} %{message}", "TestLogCallpath.String.rec...a.b.c")
	testCallpath(t, 1, "%{callpath:-1} %{message}", "TestLogCallpath.String.rec...a.b.c")
	testCallpath(t, 2, "%{callpath:0} %{message}", "TestLogCallpath.String.rec...a.b.c")
	testCallpath(t, 3, "%{callpath:1} %{message}", "~.c")
	testCallpath(t, 4, "%{callpath:2} %{message}", "~.b.c")
	testCallpath(t, 5, "%{callpath:3} %{message}", "~.a.b.c")
	//testCallpath(t, 6, `%{time:0102 15:04:05.000} %{callpath:3} - %{level:.4s} %{id:03x} %{message}`, "~.a.b.c")
}

func BenchmarkLogMemoryBackendIgnored(b *testing.B) {
	backend := SetBackend(NewMemoryBackend(1024))
	backend.SetLevel(INFO, "")
	RunLogBenchmark(b)
}

func BenchmarkLogMemoryBackend(b *testing.B) {
	backend := SetBackend(NewMemoryBackend(1024))
	backend.SetLevel(DEBUG, "")
	RunLogBenchmark(b)
}

func BenchmarkLogChannelMemoryBackend(b *testing.B) {
	channelBackend := NewChannelMemoryBackend(1024)
	backend := SetBackend(channelBackend)
	backend.SetLevel(DEBUG, "")
	RunLogBenchmark(b)
	channelBackend.Flush()
}

func BenchmarkLogLeveled(b *testing.B) {
	backend := SetBackend(NewLogBackend(ioutil.Discard, "", 0))
	backend.SetLevel(INFO, "")

	RunLogBenchmark(b)
}

func BenchmarkLogLogBackend(b *testing.B) {
	backend := SetBackend(NewLogBackend(ioutil.Discard, "", 0))
	backend.SetLevel(DEBUG, "")
	RunLogBenchmark(b)
}

func BenchmarkLogLogBackendColor(b *testing.B) {
	colorizer := NewLogBackend(ioutil.Discard, "", 0)
	colorizer.Color = true
	backend := SetBackend(colorizer)
	backend.SetLevel(DEBUG, "")
	RunLogBenchmark(b)
}

func BenchmarkLogLogBackendStdFlags(b *testing.B) {
	backend := SetBackend(NewLogBackend(ioutil.Discard, "", log.LstdFlags))
	backend.SetLevel(DEBUG, "")
	RunLogBenchmark(b)
}

func BenchmarkLogLogBackendLongFileFlag(b *testing.B) {
	backend := SetBackend(NewLogBackend(ioutil.Discard, "", log.Llongfile))
	backend.SetLevel(DEBUG, "")
	RunLogBenchmark(b)
}

func RunLogBenchmark(b *testing.B) {
	password := Password("foo")
	log := MustGetLogger("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Debug("log line for %d and this is rectified: %s", i, password)
	}
}

func BenchmarkLogFixed(b *testing.B) {
	backend := SetBackend(NewLogBackend(ioutil.Discard, "", 0))
	backend.SetLevel(DEBUG, "")

	RunLogBenchmarkFixedString(b)
}

func BenchmarkLogFixedIgnored(b *testing.B) {
	backend := SetBackend(NewLogBackend(ioutil.Discard, "", 0))
	backend.SetLevel(INFO, "")
	RunLogBenchmarkFixedString(b)
}

func RunLogBenchmarkFixedString(b *testing.B) {
	log := MustGetLogger("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Debug("some random fixed text")
	}
}
