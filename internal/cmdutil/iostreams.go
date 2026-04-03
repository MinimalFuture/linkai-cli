package cmdutil

import "io"

type IOStreams struct {
	In              io.Reader
	Out             io.Writer
	ErrOut          io.Writer
	IsTerminal      bool // true when stdout is a terminal
	IsStdinTerminal bool // true when stdin is a terminal
}
