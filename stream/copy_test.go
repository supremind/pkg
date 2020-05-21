package stream

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type slowWriter struct {
	delay time.Duration
	w     io.Writer
}

func (sw *slowWriter) Write(p []byte) (n int, e error) {
	time.Sleep(sw.delay)
	return sw.w.Write(p)
}

type slowReader struct {
	delay time.Duration
	r     io.Reader
}

func (sr *slowReader) Read(p []byte) (n int, e error) {
	time.Sleep(sr.delay)
	return sr.r.Read(p)
}

func TestDoubleBufferedCopy(t *testing.T) {
	content := bytes.Repeat([]byte("0123456789abcdef"), 16)
	input := bytes.NewBuffer(content)
	output := &bytes.Buffer{}

	// 15 is in purpose, buf will not be full in the last run
	n, e := DoubleBufferedCopy(output, input, 15)
	assert.NoError(t, e)
	assert.Equal(t, int64(256), n)
	assert.Equal(t, content, output.Bytes())
}

/*
BenchmarkDoubleBufferedCopy-4                 13          91187478 ns/op            4841 B/op         27 allocs/op
BenchmarkStdCopy-4                             6         178568924 ns/op            3578 B/op          8 allocs/op

❯ go test -benchmem -benchtime=60s -bench=.
goos: linux
goarch: amd64
pkg: github.com/supremind/pkg/stream
BenchmarkDoubleBufferedCopy-8                 69        1029163522 ns/op            4624 B/op         25 allocs/op
BenchmarkStdCopy-8                            34        2009745194 ns/op            3585 B/op          9 allocs/op
*/

func BenchmarkDoubleBufferedCopy(b *testing.B) {
	input := &slowReader{
		delay: time.Millisecond,
	}
	output := &slowWriter{
		delay: time.Millisecond,
	}

	var n int64
	var e error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.r = bytes.NewBuffer(make([]byte, 1024))
		output.w = &bytes.Buffer{}
		n, e = DoubleBufferedCopy(output, input, 16)
	}
	if e != nil {
		b.Fatal(e)
	}
	if n != 1024 {
		b.Fatal(n)
	}
}

func BenchmarkStdCopy(b *testing.B) {
	input := &slowReader{
		delay: time.Millisecond,
	}
	output := &slowWriter{
		delay: time.Millisecond,
	}

	var n int64
	var e error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.r = bytes.NewBuffer(make([]byte, 1024))
		output.w = &bytes.Buffer{}

		// runs same trips as DoubleBufferedCopy, allocates less memory
		cpBuf := make([]byte, 16)
		n, e = io.CopyBuffer(output, input, cpBuf)
	}
	if e != nil {
		b.Fatal(e)
	}
	if n != 1024 {
		b.Fatal(n)
	}
}
