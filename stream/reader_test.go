package stream

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneReader(t *testing.T) {
	{
		r := strings.NewReader("12345")
		r1, r2 := CloneReader(r)

		assert.NoError(t, checkReader("12345", r1))
		assert.NoError(t, checkReader("12345", r2))
	}
	{
		r := strings.NewReader("12345")
		r1, r2 := CloneReader(r)

		_ = r1
		assert.NoError(t, checkReader("12345", r2))
	}
	{
		// this is the main difference compared with io.TeeReader

		r := strings.NewReader("12345")
		r1, r2 := CloneReader(r)

		var buf bytes.Buffer
		_, e := io.CopyN(&buf, r1, 2)
		assert.NoError(t, e)

		assert.NoError(t, checkReader("12", &buf))
		assert.NoError(t, checkReader("12345", r2))
	}
}

func TestTeeReader(t *testing.T) {
	r := strings.NewReader("12345")
	var rw1 bytes.Buffer
	r2 := io.TeeReader(r, &rw1)

	var buf bytes.Buffer
	_, e := io.CopyN(&buf, r2, 2)
	assert.NoError(t, e)

	assert.NoError(t, checkReader("12", &buf))
	assert.NoError(t, checkReader("12", &rw1))
}

func checkReader(expected string, r io.Reader) error {
	b, e := ioutil.ReadAll(r)
	if e != nil {
		return e
	}

	if string(b) != expected {
		return fmt.Errorf("expected %s, got %s", expected, string(b))
	}

	return nil
}
