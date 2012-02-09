include $(GOROOT)/src/Make.inc

TARG=ed

GOFILES=\
	ed.go\
	buf.go\
	cmd.go\
	op.go\

include $(GOROOT)/src/Make.cmd
