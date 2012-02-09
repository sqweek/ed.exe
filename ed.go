package ed

import (
	"os"
	/*"fmt"*/
	"bufio"
	"flag"
)

type Cmd struct {
	op Op
	r Range
}

func (c *Cmd) Run(b *Buffer) {
	c.op.Run(b, c.r)
}

func EdError(err string) {
	os.Stderr.WriteString("?  " + err + "\n")
}

var stdin *bufio.Reader

func EdInput() (string, bool) {
	line, _, e := stdin.ReadLine()
	if e != nil {
		return "", true
	}
	return string(line), false
}

func main() {
	var b Buffer
	var ctx ParseContext
	flag.Parse()
	if flag.NArg() == 1 {
		var cmd Cmd
		cmd.r.line0, cmd.r.lineN = 0, 0
		cmd.op = NewReadOp(flag.Arg(0))
		cmd.Run(&b)
	}

	stdin = bufio.NewReader(os.Stdin)
	for {
		line, done := EdInput()
		if done {
			break
		}
		var cmd Cmd
		cmd.r, cmd.op = CmdParse(line, &b, &ctx)
		cmd.Run(&b)
		_, quit := cmd.op.(*QuitOp)
		if(quit) {
			break
		}
	}
}
