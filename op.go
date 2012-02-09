package ed

import (
	"regexp"
	"os"
	"bufio"
	"strconv"
)

const (
	LINE_NOP = iota
	LINE_DEL
	LINE_REPLACE
)

const (
	INPUT_INSERT = iota
	INPUT_CHANGE
	INPUT_APPEND
)

type Op interface {
	Run(*Buffer, Range)
}

type LineOpFn func(int, string) (int, string)

type LineOp struct {
	fn LineOpFn
}

func MkLineOp(fn LineOpFn) Op {
	var o LineOp
	o.fn = fn
	return Op(&o)
}

func MkOpFnErr(err string) LineOpFn {
	done := false
	return func(linenum int, line string) (int, string) {
		if done == false {
			EdError(err + "\n")
			done = true
		}
		return LINE_NOP, ""
	}
}

func (op *LineOp) Run(buf *Buffer, r Range) {
	for i:=r.line0; i<=r.lineN; i++ {
		act, str := op.fn(i, buf.GetLine(i))
		switch {
			case act == LINE_REPLACE:
				buf.SetLine(i, str)
			case act == LINE_NOP:
				/* do nothing */
			case act == LINE_DEL:
				/* TODO */
		}
	}
}


/* n	- display line number and line contents */
func OpFnPrintN(linenum int, line string) (int, string) {
	os.Stdout.WriteString(strconv.Itoa(linenum+1) + "\t" + line + "\n")
	return LINE_NOP, ""
}


/* p	- display line contents */
func OpFnPrint(linenum int, line string) (int, string) {
	os.Stdout.WriteString(line + "\n")
	return LINE_NOP, ""
}

/* s	- substitute text */
type Subst struct {
	re regexp.Regexp
}

func MkOpFnSubst(re *regexp.Regexp, repl string, opts string) LineOpFn {
	/* TODO opts g123456789 */
	var flagP, flagG bool
	var flagI [9]bool
	for i := 0; i < len(opts); i++ {
		switch c := opts[i]; true {
			case c == 'p':
				flagP = true
			case c == 'g':
				flagG = true
			case c >= '1' && c <= '9':
				flagI[c - '1'] = true
		}
	}
	return func(linenum int, line string) (int, string) {
		matches := re.FindAllStringIndex(line, -1)
		if matches == nil {
			EdError("no match")
			return LINE_NOP, ""
		}
		result := line[0:matches[0][0]]
		for i := 0; i < len(matches); i++ {
			if i > 0 {
				result += line[matches[i-1][1]:matches[i][0]]
			}
			for j := 0; j < len(matches[i]); j++ {
				print(matches[i][j], " ")
			}
			println()
			if flagG || (i <= 9 && flagI[i]) {
				result += repl
			} else {
				result += line[matches[i][0]:matches[i][1]]
			}
		}
		result += line[matches[len(matches)-1][1]:]
		if flagP {
			os.Stdout.WriteString(result + "\n")
		}
		return LINE_REPLACE, result
	}
}


/* g    - apply a LineOp to lines matching a regex */
func MkOpFnGlob(re *regexp.Regexp, op *LineOp) LineOpFn {
	return func(linenum int, line string) (int, string) {
		if re.MatchString(line) {
			return op.fn(linenum, line)
		}
		return LINE_NOP, ""
	}
}


/* r	- read file into working buffer */
type ReadOp struct {
	filename string
}

func NewReadOp(filename string) *ReadOp {
	o := new(ReadOp)
	o.filename = filename
	return o
}

func (op *ReadOp) Run(buf *Buffer, r Range) {
	/* TODO trigger error on invalid range */
	f, err := os.Open(op.filename)
	if err != nil {
		EdError(op.filename + ": " + err.String())
		return
	}
	rdr := bufio.NewReader(f)
	err = buf.Read(rdr, op.filename)
	if err != nil {
		EdError(op.filename + ": " + err.String())
	}
	f.Close()
}


/* w - write buffer to file */
type WriteOp struct {
	filename string
}

func NewWriteOp(filename string) *WriteOp {
	var op WriteOp
	op.filename = filename
	return &op
}

func (op *WriteOp) Run(buf *Buffer, r Range) {
	f, err := os.Create(op.filename)
	if err != nil {
		EdError(op.filename + ": " + err.String())
		return
	}
	writer := bufio.NewWriter(f)
	err = buf.Write(writer)
	if err != nil {
		EdError(op.filename + ": " + err.String())
	}
	f.Close()
}

/* q	- quit */
type QuitOp struct {}

func (op *QuitOp) Run(buf *Buffer, r Range) {
	/* do nothing */
}


/* no-op, used when error has occured */
type Nop struct {}

func (op *Nop) Run(buf *Buffer, r Range) {
	/* do nothing */
}


/* InputOp is used for ops that access user input - a, i, c */
type InputOp struct {
	OpType int
}

func NewInputOp(t int) *InputOp {
	var op InputOp
	op.OpType = t
	return &op
}

func (op *InputOp) Run(buf *Buffer, r Range) {
	lines := make([]string, 0, 32)
	for {
		line, done := EdInput()
		if line == "." || done {
			break
		}
		lines = append(lines, line)
	}
	var insPoint int
	switch {
		case op.OpType == INPUT_CHANGE:
			buf.DeleteLines(r)
			insPoint = r.line0
		case op.OpType == INPUT_INSERT:
			insPoint = r.line0
		case op.OpType == INPUT_APPEND:
			insPoint = r.lineN + 1
			if buf.NumLines() == 0 && r.lineN == 0 {
				insPoint = 0
			}
		default:
			panic("unknown INPUT OpType")
	}
	buf.InsertLines(lines, insPoint)
}