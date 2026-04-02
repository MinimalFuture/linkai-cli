package cmdutil

import "io"

type IOStreams struct {
	In         io.Reader
	Out        io.Writer
	ErrOut     io.Writer
	IsTerminal bool
}
