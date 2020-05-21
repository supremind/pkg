package stream

import (
	"context"
	"fmt"
	"io"
	"sync"

	"golang.org/x/sync/errgroup"
)

// DoubleBufferedCopy could be helpful when reader and writer are both slow.
// It buffers reading and writing simultaneously, and exchanges double buffers between reader and writer
func DoubleBufferedCopy(w io.Writer, r io.Reader, bufSize int) (n int64, e error) {
	var rn, wn int64

	empty := make(chan []byte, 2)
	full := make(chan []byte)

	eg, ctx := errgroup.WithContext(context.TODO())

	eg.Go(func() error {
		b1 := make([]byte, bufSize)
		b2 := make([]byte, bufSize)

		empty <- b1
		empty <- b2

		return nil
	})

	eg.Go(func() error {
		defer close(full)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()

			case buf := <-empty:
				buf = buf[:bufSize]
				n, e := r.Read(buf)
				rn += int64(n)
				buf = buf[:n]
				full <- buf

				if e != nil {
					if e == io.EOF {
						return nil
					}
					return e
				}
			}
		}
	})

	eg.Go(func() error {
		// write as many as it could, ignore ctx.Done
		for buf := range full {
			if len(buf) == 0 {
				empty <- buf
				continue
			}
			n, e := w.Write(buf)
			wn += int64(n)

			empty <- buf
			if e != nil {
				return e
			}
			if n != len(buf) {
				return io.ErrShortWrite
			}
		}

		return nil
	})

	e = eg.Wait()
	<-empty
	<-empty
	close(empty)

	n = wn
	if e == nil && wn < rn {
		e = io.ErrShortWrite
	}

	return
}

func CopyInBlocks(ctx context.Context, wa io.WriterAt, ra io.ReaderAt, size, blockSize int64, workers int) (n int64, e error) {
	if size <= 0 {
		return 0, fmt.Errorf("invalid size: %d", size)
	}
	if blockSize <= 0 {
		return 0, fmt.Errorf("invalid block size: %d", blockSize)
	}
	if workers <= 0 {
		return 0, fmt.Errorf("invalid number of workers: %d", workers)
	}

	type copyJob struct {
		from, to int64
		buf      []byte
	}

	// half of the workers always read, others write
	bufs := make(chan []byte, workers)
	rJobs := make(chan copyJob)
	wJobs := make(chan copyJob)

	eg, ctx := errgroup.WithContext(ctx)

	// init buffers, will not block
	for i := 0; i < workers; i++ {
		bufs <- make([]byte, int(blockSize))
	}

	eg.Go(func() error {
		defer close(rJobs)
		var from, to int64
		for {
			if from >= size {
				return nil
			}
			to = from + blockSize
			if to > size {
				to = size
			}

			buf := <-bufs
			select {
			case rJobs <- copyJob{from: from, to: to, buf: buf}:
				from = to

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		eg.Go(func() error {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()

				case job, ok := <-rJobs:
					if !ok {
						return nil
					}

					buf := job.buf[:int(blockSize)]
					r := io.NewSectionReader(ra, job.from, job.to-job.from)

					var n, rd int
					var e error
					for e == nil && n < int(job.to-job.from) {
						rd, e = r.Read(buf[n:])
						n += rd
					}
					if e != nil && e != io.EOF {
						return fmt.Errorf("reading block %d-%d: %w", job.from, job.to, e)
					}

					job.buf = buf[:n]
					wJobs <- job
				}
			}
		})
	}

	eg.Go(func() error {
		wg.Wait()
		close(wJobs)
		return nil
	})

	for i := 0; i < workers; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()

				case job, ok := <-wJobs:
					if !ok {
						return nil
					}

					w := NewSegmentWriter(wa, job.from, job.to-job.from)
					buf := job.buf

					var n, wr int
					var e error
					for e == nil && n < int(job.to-job.from) {
						wr, e = w.Write(buf[n:])
						n += wr
					}
					if e != nil {
						return fmt.Errorf("writing block %d-%d: %w", job.from, job.to, e)
					}

					bufs <- buf
				}
			}
		})
	}

	e = eg.Wait()
	if e == nil {
		n = size
	}

	for i := 0; i < workers; i++ {
		<-bufs
	}
	return
}
