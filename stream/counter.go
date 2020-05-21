package stream

import (
	"io"
	"sync/atomic"
)

type Counter interface {
	Count() <-chan int64
	io.Closer
}

type bytesCounter struct {
	cnt      chan int64
	consumed int64
}

func newBytesCounter() *bytesCounter {
	return &bytesCounter{
		cnt: make(chan int64, 1),
	}
}

func (c *bytesCounter) Count() <-chan int64 {
	return c.cnt
}

func (c *bytesCounter) Close() error {
	close(c.cnt)
	return nil
}

func (c *bytesCounter) Accumulate(n int64) {
	c.cnt <- int64(atomic.AddInt64(&c.consumed, n))
}

// CountingReader counts bytes read from the source reader
type CountingReader struct {
	src io.Reader
	*bytesCounter
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{
		src:          r,
		bytesCounter: newBytesCounter(),
	}
}

func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.src.Read(p)
	cr.bytesCounter.Accumulate(int64(n))
	return
}

func (cr *CountingReader) Close() error {
	if c, ok := cr.src.(io.Closer); ok {
		c.Close()
	}
	return cr.bytesCounter.Close()
}

// CountingReaderAt counts bytes read from the source ReaderAt
type CountingReaderAt struct {
	src io.ReaderAt
	*bytesCounter
}

func NewCountingReaderAt(r io.ReaderAt) *CountingReaderAt {
	return &CountingReaderAt{
		src:          r,
		bytesCounter: newBytesCounter(),
	}
}

func (cr *CountingReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = cr.src.ReadAt(p, off)
	cr.bytesCounter.Accumulate(int64(n))
	return
}

func (cr *CountingReaderAt) Close() error {
	if c, ok := cr.src.(io.Closer); ok {
		c.Close()
	}
	return cr.bytesCounter.Close()
}

// CountingWriter counts bytes written in destination writer
type CountingWriter struct {
	dst io.Writer
	*bytesCounter
}

func NewCountingWriter(dst io.Writer) *CountingWriter {
	return &CountingWriter{
		dst:          dst,
		bytesCounter: newBytesCounter(),
	}
}

func (cw *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.dst.Write(p)
	cw.bytesCounter.Accumulate(int64(n))
	return
}

func (cw *CountingWriter) Close() error {
	if c, ok := cw.dst.(io.Closer); ok {
		c.Close()
	}
	return cw.bytesCounter.Close()
}

// CountintWriterAt counts bytes written in destination WriterAt
type CountingWriterAt struct {
	dst io.WriterAt
	*bytesCounter
}

func NewCountingWriterAt(dst io.WriterAt) *CountingWriterAt {
	return &CountingWriterAt{
		dst:          dst,
		bytesCounter: newBytesCounter(),
	}
}

func (cw *CountingWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = cw.dst.WriteAt(p, off)
	cw.bytesCounter.Accumulate(int64(n))
	return
}

func (cw *CountingWriterAt) Close() error {
	if c, ok := cw.dst.(io.Closer); ok {
		c.Close()
	}
	return cw.bytesCounter.Close()
}

// WriteCounter counts bytes written in, it writes nothing
type WriteCounter struct {
	*bytesCounter
}

func NewWriteCounter() *WriteCounter {
	return &WriteCounter{newBytesCounter()}
}

func (rc *WriteCounter) Write(p []byte) (n int, err error) {
	n = len(p)
	rc.bytesCounter.Accumulate(int64(n))
	return
}

// ReadCounter counts bytes the caller wants to read, the input slice is never touched
type ReadCounter struct {
	*bytesCounter
}

func NewReadCounter() *ReadCounter {
	return &ReadCounter{newBytesCounter()}
}

func (rc *ReadCounter) Read(p []byte) (n int, err error) {
	n = len(p)
	rc.bytesCounter.Accumulate(int64(n))
	return
}

type CountingAutoReader struct {
	ar AutoReader
	*bytesCounter
}

func NewCountingAutoReader(ar AutoReader) *CountingAutoReader {
	return &CountingAutoReader{
		ar:           ar,
		bytesCounter: newBytesCounter(),
	}
}

func (ar *CountingAutoReader) Read(p []byte) (n int, e error) {
	n, e = ar.ar.Read(p)
	ar.bytesCounter.Accumulate(int64(n))
	return
}

func (ar *CountingAutoReader) ReadAt(p []byte, off int64) (n int, e error) {
	n, e = ar.ar.ReadAt(p, off)
	ar.bytesCounter.Accumulate(int64(n))
	return
}

func (ar *CountingAutoReader) Close() error {
	if c, ok := ar.ar.(io.Closer); ok {
		c.Close()
	}
	return ar.bytesCounter.Close()
}

type CountingAutoWriter struct {
	aw AutoWriter
	*bytesCounter
}

func NewCountingAutoWriter(aw AutoWriter) *CountingAutoWriter {
	return &CountingAutoWriter{
		aw:           aw,
		bytesCounter: newBytesCounter(),
	}
}

func (aw *CountingAutoWriter) Write(p []byte) (n int, e error) {
	n, e = aw.aw.Write(p)
	aw.bytesCounter.Accumulate(int64(n))
	return
}

func (aw *CountingAutoWriter) WriteAt(p []byte, off int64) (n int, e error) {
	n, e = aw.aw.WriteAt(p, off)
	aw.bytesCounter.Accumulate(int64(n))
	return
}

func (aw *CountingAutoWriter) Close() error {
	if c, ok := aw.aw.(io.Closer); ok {
		c.Close()
	}
	return aw.bytesCounter.Close()
}
