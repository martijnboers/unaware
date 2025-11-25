package test

import "io"

// errorReader is a helper for testing that simulates an error during reading.
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
