package ed

import (
	"regexp"
	"os"
	"bufio"
	"strconv"
)

const (
	INPUT_INSERT = iota
	INPUT_CHANGE
	INPUT_APPEND
)

type Op interface {
	Run(*Buffer, Range)
}

type LineOpFn func(int, *string) *string

type LineOp struct {
	fn LineOpFn
}

func MkLineOp(fn LineOpFn) Op {
	var o LineOp
	o.fn = fn
	return Op(&o)
}

func (op *LineOp) Run(buf *Buffer, r Range) {
	deleted := make([]int, 0)
	for i:=r.line0; i<=r.lineN; i++ {
		line := buf.GetLine(i)
		newLine := op.fn(i, &line)
		if newLine != nil {
			if newLine != &line {
				buf.SetLine(i, *newLine)
				buf.dot = i
			}
		} else {
			deleted = append(deleted, i)
			buf.dot = i
		}
	}
	for i:=len(deleted)-1; i>=0; i-- {
		_ = buf.DeleteLines(Range{deleted[i], deleted[i]})
	}
	if buf.dot >= buf.NumLines() {
		buf.dot = buf.NumLines()-1
	}
}


/* n	- display line number and line contents */
func OpFnPrintN(linenum int, line *string) *string {
	os.Stdout.WriteString(strconv.Itoa(linenum+1) + "\t" + *line + "\n")
	return line
}


/* p	- display line contents */
func OpFnPrint(linenum int, line *string) *string {
	os.Stdout.WriteString(*line + "\n")
	return line
}


/* d	- delete lines */
func OpFnDel(linenum int, line *string) *string {
	return nil
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
	return func(linenum int, line *string) *string {
		matches := re.FindAllStringIndex(*line, -1)
		if matches == nil {
			EdError("no match")
			return line
		}
		result := (*line)[0:matches[0][0]]
		for i := 0; i < len(matches); i++ {
			if i > 0 {
				result += (*line)[matches[i-1][1]:matches[i][0]]
			}
			/*for j := 0; j < len(matches[i]); j++ {
				print(matches[i][j], " ")
			}
			println()*/
			if flagG || (i <= 9 && flagI[i]) {
				result += repl
			} else {
				result += (*line)[matches[i][0]:matches[i][1]]
			}
		}
		result += (*line)[matches[len(matches)-1][1]:]
		if flagP {
			os.Stdout.WriteString(result + "\n")
		}
		return &result
	}
}


/* g,v  - apply a LineOp to lines matching/not matching a regex */
func MkOpFnGlob(buf *Buffer, re *regexp.Regexp, op *LineOp, matchTarget bool) LineOpFn {
	return func(linenum int, line *string) *string {
		if re.MatchString(*line) == matchTarget {
			return op.fn(linenum, line)
		}
		return line
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
			_ = buf.DeleteLines(r)
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


/* m	- move lines around */
type MoveOp struct {
	destinationLine int
}

func NewMoveOp(dest int) *MoveOp {
	var op MoveOp
	op.destinationLine = dest
	return &op
}

func (op *MoveOp) Run(buf *Buffer, r Range) {
	d := op.destinationLine
	if d < 0 || d > buf.NumLines()-1 || (d >= r.line0 && d <= r.lineN) {
		EdError("bad move target")
		return
	}
	tmp := buf.DeleteLines(r)
	if d > r.lineN {
		d -= (r.lineN - r.line0) + 1
	}
	buf.InsertLines(tmp, d)
}
