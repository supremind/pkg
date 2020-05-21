package stream

type NoErrorCloser func()

func (f NoErrorCloser) Close() error {
	f()
	return nil
}

type FuncCloser func() error

func (f FuncCloser) Close() error {
	return f()
}
