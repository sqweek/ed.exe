package main

import (
	"fmt"
	"bufio"
	"./cmd"
	"./buf"
)

type Cmd {
	op Op
	addr Address
}

func ed(b buf.Buffer, ctl chan cmd.Cmd) {
	for {
		c := <-ctl
		if c.op.OpType() == op.OpTypeLine {
			b.Apply(c.addr, c.op)
		}
	}
}

func main() {
	rdr := bufio.NewReader(os.Stdin)
	for line := rdr.ReadLine(); line != nil; line := rdr.ReadLine() {
		addr, cmd := parse(line)
		/* parse command */
		/* match address */
		/* run command on each matching line */
	}
}
