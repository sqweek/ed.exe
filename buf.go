package buf

import (
	"bufio"
	"os"
	"./cmd"
	"./op"
)

type Addr interface {
	matches(int, string) bool
}

type Buffer {
	dirty bool
	length int
	lines[] string
}

func (buf *Buffer) IsDirty() bool {
	return buf.dirty
}

func (buf *Buffer) GetLine(i int) string {
	return buf.lines[i]
}

func (buf *Buffer) SetLine(i int, str string) {
	buf.lines[i] = str
}

/* Reads all the lines available from the reader and adds them to the buffer */
func (buf *Buffer) Append(rdr bufio.Reader) os.Error {
	var prefix := ""
	for {
		segment, isPrefix, err = rdr.ReadLine()
		if err == EOF {
			return
		} else if err != nil {
			return err
		}
		if isPrefix {
			prefix = prefix + segment
		} else {
			append(buf.lines, prefix + segment)
			prefix = ""
		}
	}
}

func (buf *Buffer) Apply(addr cmd.Addr, o op.LineOp) {
	for i := 0; i < length(buf.lines); i++ {
		if(addr.matches(i, buf.lines[i])) {
			res, del := o.Run(buf.lines[i])
		}
	}
}
