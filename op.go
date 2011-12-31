package op

const (
	LineType = iota,
	BufType,
	QuitType
)

type Op interface {
	OpType() int
}

type LineOp interface {
	Run(string) string, bool
}

type BufOp interface {
	Run(*Buffer) *Buffer
}


/* n	- display line number and line contents */
type PrintN {}

func (p *PrintN) Run(buf Buffer, linenum int) {
	os.Stdout.WriteString(string(linenum) + "\t")
	os.Stdout.WriteString(buf.GetLine(linenum))
	os.Stdout.WriteString("\n")
}

func (p *PrintN) OpType() int {
	return LineType
}


/* p	- display line contents */
type Print {}

func (p *Print) Run(buf Buffer, linenum int) {
	os.Stdout.WriteString(buf.GetLine(linenum))
	os.Stdout.WriteString("\n")
}

func (p *Print) OpType() int {
	return LineType
}


/* s	- substitute text */
type Subst {
	re Regexp
}

func (s *Subst) Run(buf Buffer, linenum int) {
	buf.SetLine(linenum, s.re.ReplaceAll(buf.GetLine(linenum))
}

func (s *Subst) OpType() int {
	return LineType
}


/* r	- read file into working buffer */
type Read {
	filename string
}

func (op *Read) Run(buf *Buffer) {
	f, err := os.Open(c.filename)
	if err {
		Fmt.Fprintln(os.Stderr, "! ", op.filename, ":", err)
		return
	}
	rdr := bufio.NewReader(f)
	if err = buf.Append(rdr) {
		Fmt.Fprintln(os.Stderr, "! ", op.filename, ":", err)
	}
	f.close(f)
}

func (op *Read) OpType() int {
	return BufType
}


/* q	- quit */
type Quit {}

func (op *Quit) OpType() int {
	return QuitType
}
