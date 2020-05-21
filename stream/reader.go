package stream

import (
	"bufio"
	"bytes"
	"io"
)

func CloneReader(r io.Reader) (r1, r2 io.Reader) {
	var buf bytes.Buffer
	t := io.TeeReader(r, &buf)
	return t, &ReaderFork{upstream: r, buf: &buf}
}

type ReaderFork struct {
	upstream io.Reader
	buf      *bytes.Buffer
	empty    bool
}

func (r *ReaderFork) Read(p []byte) (n int, err error) {
	if !r.empty {
		n, err = r.buf.Read(p)
		if err == io.EOF {
			r.empty = true
			err = nil
		}
		return
	}
	return r.upstream.Read(p)
}

func NewSegmentReader(ra io.ReaderAt, off int64, size int64) *io.SectionReader {
	return io.NewSectionReader(ra, off, size)
}

type AutoReader interface {
	io.Reader
	io.ReaderAt
}

// BufferedAutoReader buffers calls to Read, but not ReadAt
type BufferedAutoReader struct {
	AutoReader
	buf *bufio.Reader
}

func NewBufferedAutoReader(ar AutoReader, size int) *BufferedAutoReader {
	return &BufferedAutoReader{
		AutoReader: ar,
		buf:        bufio.NewReaderSize(ar, size),
	}
}

func (ar *BufferedAutoReader) Read(p []byte) (n int, e error) {
	return ar.buf.Read(p)
}

type ReadAtCloser interface {
	io.ReaderAt
	io.Closer
}

type nopReadAtCloser struct {
	io.ReaderAt
}

func (c nopReadAtCloser) Close() error {
	return nil
}

func NopReadAtCloser(ra io.ReaderAt) ReadAtCloser {
	return nopReadAtCloser{ra}
}
