package ed

import (
	"bufio"
	"regexp"
	"os"
	//"fmt"
)


type Range struct {
	line0 int
	lineN int
}

func MkRange(line0 int, lineN int) Range {
	var r Range
	r.line0 = line0
	r.lineN = lineN
	return r
}

type Buffer struct {
	filename string
	dirty bool
	length int
	lines[] string

	dot int /* "current" line */
}

func (buf *Buffer) IsDirty() bool {
	return buf.dirty
}

func (buf *Buffer) Dot() int {
	return buf.dot
}

func (buf *Buffer) SearchForward(re *regexp.Regexp, startline int) int {
	for i:=startline+1; i<len(buf.lines); i++ {
		if re.MatchString(buf.lines[i]) {
			return i
		}
	}
	return -1
}

func (buf *Buffer) SearchBackward(re *regexp.Regexp, startline int) int {
	for i:=startline-1; i>=0; i-- {
		if re.MatchString(buf.lines[i]) {
			return i
		}
	}
	return -1
}

func (buf *Buffer) NumLines() int {
	return len(buf.lines)
}

func (buf *Buffer) GetLine(i int) string {
	buf.dot = i
	return buf.lines[i]
}

func (buf *Buffer) SetLine(i int, str string) {
	buf.dot = i
	buf.lines[i] = str
	buf.dirty = true
}

/* Reads all the lines available from the reader and adds them to the buffer */
func (buf *Buffer) Read(rdr *bufio.Reader, filename string) os.Error {
	prefix := ""
	for {
		segment, isPrefix, err := rdr.ReadLine()
		if err == os.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if isPrefix {
			prefix = prefix + string(segment)
		} else {
			buf.lines = append(buf.lines, prefix + string(segment))
			prefix = ""
		}
	}
	buf.filename = filename
	buf.dot = len(buf.lines) - 1
	buf.dirty = false
	return nil
}

func (buf *Buffer) Write(writer *bufio.Writer) os.Error {
	var err os.Error
	for i := 0; i < len(buf.lines); i++ {
		/* TODO try to write remaining bytes on error */
		_, e := writer.WriteString(buf.lines[i]+"\n")
		if e != nil && err != nil {
			err = e
		}
	}
	writer.Flush()
	buf.dirty = false
	return err
}

func (buf *Buffer) DeleteLines(r Range) {
	n := r.lineN - r.line0 + 1
	copy(buf.lines[r.line0:len(buf.lines)-n], buf.lines[r.lineN+1:])
	buf.lines = buf.lines[:len(buf.lines)-n]
	buf.dirty = true
}

func (buf *Buffer) InsertLines(newLines []string, insPoint int) {
	/* XXX surely there's a better way to grow a slice than this */
	n := len(newLines)
	if n == 0 {
		return
	}
	//EdError(fmt.Sprint("InsertLines: got ", n, " lines for ", insPoint))
	for i := 0; i < n; i++ {
		//EdError(fmt.Sprint(i, newLines[i]))
		buf.lines = append(buf.lines, "")
	}
	copy(buf.lines[insPoint+n:], buf.lines[insPoint:len(buf.lines)-n])
	copy(buf.lines[insPoint:insPoint+n], newLines)
	buf.dirty = true
}

