package stream

import (
	"bufio"
	"io"
)

type SegmentWriter struct {
	wa   io.WriterAt
	off  int64
	size int64
}

// should not be called in parallel
func (sw *SegmentWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	if int64(n) > sw.size {
		n = int(sw.size)
	}
	n, err = sw.wa.WriteAt(p[:n], sw.off)
	sw.off += int64(n)
	sw.size -= int64(n)
	return
}

func NewSegmentWriter(wa io.WriterAt, off int64, size int64) *SegmentWriter {
	return &SegmentWriter{wa: wa, off: off, size: size}
}

type AutoWriter interface {
	io.Writer
	io.WriterAt
}

type BufferedAutoWriter struct {
	AutoWriter
	buf *bufio.Writer
}

func NewBufferedAutoWriter(aw AutoWriter, size int) *BufferedAutoWriter {
	return &BufferedAutoWriter{
		AutoWriter: aw,
		buf:        bufio.NewWriterSize(aw, size),
	}
}

func (aw *BufferedAutoWriter) Write(p []byte) (n int, e error) {
	return aw.buf.Write(p)
}

func (aw *BufferedAutoWriter) Close() error {
	return aw.buf.Flush()
}
