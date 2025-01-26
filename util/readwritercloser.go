package util

import "io"

type ReadWriteCloser struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

func (rw *ReadWriteCloser) Read(p []byte) (n int, err error) {
	return rw.Reader.Read(p)
}

func (rw *ReadWriteCloser) Write(p []byte) (n int, err error) {
	return rw.Writer.Write(p)
}

func (rw *ReadWriteCloser) Close() error {
	err := rw.Reader.Close()
	err2 := rw.Writer.Close()

	if err != nil {
		return err
	} else {
		return err2
	}
}
