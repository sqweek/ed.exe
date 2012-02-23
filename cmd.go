package ed

import (
	"unicode"
	"strconv"
	"regexp"
	"strings"
	"fmt"
	"os"
)

type ParseContext struct {
	buf *Buffer

	prevState []string /* used for undo */

	lastSearch *regexp.Regexp
	lastReplace *string
	lastSubstOpts *string
}

func CmdParse(str string, ctx *ParseContext) (Range, Op) {
	var oper Op
	var nop Nop
	addr, rest := parseRange(str, ctx)
	if addr.line0 == -1 || addr.lineN == -1 {
		EdError("no match")
		oper = &nop
	} else if addr.line0 < 0 ||
	  ((ctx.buf.NumLines() > 0) && addr.lineN > ctx.buf.NumLines()-1) ||
	  ((ctx.buf.NumLines() == 0) && addr.lineN > 0) {
		EdError("bad range")
		oper = &nop
	} else {
		oper, _ = parseOp(rest, ctx)
	}
	if(len(rest) == len(str) && rest[0] == 'g') {
		/* XXX hack! */
		/* no range was explicitly specified on g command, default to whole buffer */
		addr.line0 = 0
		addr.lineN = ctx.buf.NumLines()-1
	}
	return addr, oper
}

func parseRange(str string, ctx *ParseContext) (Range, string) {
	var rest string
	var line0, lineN int
	line0, rest = consumeAddr(str, -1, ctx)
	//fmt.Fprintln(os.Stderr, "parseRange(\"", str, "\"):", rest, line0)
	if len(rest) > 0 && rest[0] == ',' {
		lineN, rest = consumeAddr(rest[1:], line0, ctx)
	} else {
		lineN = line0
	}
	//fmt.Fprintln(os.Stderr, "parseRange:", rest, lineN)
	return MkRange(line0, lineN), rest
}

func consumeAddr(str string, line0 int, ctx *ParseContext) (int, string) {
	var searchStart int
	var defaultLine int
	if line0 == -1 {
		searchStart = ctx.buf.Dot()
		defaultLine = ctx.buf.Dot()
	} else {
		searchStart = line0
		defaultLine = ctx.buf.NumLines()-1
	}
	if len(str) > 0 {
		switch {
			case str[0] == '/' || str[0] == '?':
				re, rest, _ := consumeRegex(str)
				if len(re) > 0 {
					regex, err := regexp.Compile(re)
					if err != nil {
						return -1, rest
					}
					ctx.lastSearch = regex
				} else if ctx.lastSearch == nil {
					return -1, rest
				}
				if str[0] == '/' {
					return ctx.buf.SearchForward(ctx.lastSearch, searchStart), rest
				} else {
					return ctx.buf.SearchBackward(ctx.lastSearch, searchStart), rest
				}
			case str[0] == '^':
				return 0, str[1:]
			case str[0] == ',' && line0 == -1:
				return 0, str
			case str[0] == '$':
				return ctx.buf.NumLines()-1, str[1:]
			case str[0] == '.':
				return ctx.buf.Dot(), str[1:]
			case unicode.IsDigit(int(str[0])):
				num, rest := consumeNumber(str)
				i, _ := strconv.Atoi(num)
				/* subtract 1 to map from user line number to internal array index */
				return i-1, rest
			case str[0] == '-' || str[0] == '+':
				num, rest := consumeNumber(str[1:])
				num = str[0:1] + num
				i, _ := strconv.Atoi(num)
				return ctx.buf.Dot() + i, rest
		}
	}
	return defaultLine, str
}

/* returns the regex string, any remaining characters, and whether a terminating delimiter was found */
func consumeRegex(str string) (string, string, bool) {
	var i int
	delim := str[0]
	for i = 1; i < len(str) && str[i] != delim; i++ {
		if str[i] == '\\' {
			i++
		}
	}
	if i >= len(str) {
		return str[1:], "", false
	}
	return str[1:i], str[i+1:], true
}

func consumeNumber(str string) (string, string) {
	var i int
	for i = 0; i < len(str) && unicode.IsDigit(int(str[i])); i++ {
	}
	return str[:i], str[i:]
}

func parseOp(str string, ctx *ParseContext) (Op, string) {
	var o Op
	var rest string
	if len(str) < 1 {
		str = "p"
	}
	switch c := str[0]; true {
		case c == 'n':
			o = MkLineOp(OpFnPrintN)
			rest = str[1:]
		case c == 'p':
			o = MkLineOp(OpFnPrint)
			rest = str[1:]
		case c == 'd':
			o = MkLineOp(OpFnDel)
			rest = str[1:]
		case c == 'q':
			if ctx.buf.IsDirty() {
				EdError("dirty buffer")
				ctx.buf.dirty = false
				o = new(Nop)
			} else {
				o = new(QuitOp)
			}
			rest = str[1:]
		case c == 'r' || c == 'w':
			var filename string
			i := strings.Index(str, " ")
			if i == -1 {
				filename = ctx.buf.filename
			} else {
				filename = str[i+1:]
			}
			if filename == "" {
				EdError("no filename")
				o = new(Nop)
			} else {
				ctx.buf.filename = filename
				if c == 'r' {
					if ctx.buf.dirty {
						EdError("dirty buffer")
						ctx.buf.dirty = false
						o = new(Nop)
					} else {
						o = NewReadOp(filename)
					}
				} else {
					o = NewWriteOp(filename)
				}
			}
			rest = ""
		case c == 's':
			var match, replace, opts string
			var terminated, terminated2 bool
			var err os.Error
			var re *regexp.Regexp
			opts = "1p"
			if len(str) > 1 {
				delim := str[1]
				match, rest, terminated = consumeRegex(str[1:])
				if len(match) > 0 {
					re, err = regexp.Compile(match)
				}
				if err == nil {
					if re != nil {
						ctx.lastSearch = re
					}
					if terminated {
						var o string
						replace, o, terminated2 = consumeRegex(string(delim) + rest)
						if terminated2 {
							opts = o
						}
						ctx.lastReplace = &replace
					} else {
						ctx.lastReplace = ""
					}
				}
			}
			if err != nil {
				EdError("bad regex")
				o = new(Nop)
			} else if ctx.lastSearch == nil {
				EdError("no regex")
				o = new(Nop)
			} else {
				o = MkLineOp(MkOpFnSubst(ctx.lastSearch, *ctx.lastReplace, opts))
			}
			rest = ""
		case c == 'c':
			o = NewInputOp(INPUT_CHANGE)
			rest = str[1:]
		case c == 'a':
			o = NewInputOp(INPUT_APPEND)
			rest = str[1:]
		case c == 'i':
			o = NewInputOp(INPUT_INSERT)
			rest = str[1:]
		case c == 'm':
			var dest string
			dest, rest = consumeNumber(str[1:])
			if len(dest) < 0 {
				EdError("move requires target")
				o = new(Nop)
			} else {
				i, _ := strconv.Atoi(dest)
				o = NewMoveOp(i)
			}
		case c == 'g':
			search, subcmd, _ := consumeRegex(str[1:])
			if len(subcmd) < 1 {
				EdError("no command supplied for g//")
				o = new(Nop)
			} else if len(search) > 0 {
				re, err := regexp.Compile(search)
				if err != nil {
					EdError("bad regex")
					o = new(Nop)
				}
				ctx.lastSearch = re
			} else if ctx.lastSearch == nil {
				EdError("no regex")
				o = new(Nop)
			}
			if o == nil {
				var subop Op
				subop, rest = parseOp(subcmd, ctx)
				lop, isLineOp := subop.(*LineOp)
				if isLineOp {
					o = MkLineOp(MkOpFnGlob(ctx.buf, ctx.lastSearch, lop))
				} else {
					EdError(fmt.Sprintf("can't use '%c' with g//", subcmd[0]))
					o = new(Nop)
				}
			}	
		default:
			EdError(string(c) + ": unrecognised command")
			o = new(Nop)
			rest = ""
	}
	return o, rest
}
