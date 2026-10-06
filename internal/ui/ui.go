package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

var (
	UseColor = true
	Quiet    = false
)

func Init(noColor, quiet bool) {
	if noColor {
		UseColor = false
	}
	Quiet = quiet
}

func Info(format string, a ...any) {
	if Quiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	if UseColor {
		fmt.Fprintf(os.Stderr, Cyan+"[i] "+Reset+msg+"\n")
	} else {
		fmt.Fprintf(os.Stderr, "[i] "+msg+"\n")
	}
}

func Success(format string, a ...any) {
	if UseColor {
		fmt.Fprintf(os.Stderr, Green+"[+] "+Reset+format+"\n", a...)
	} else {
		fmt.Fprintf(os.Stderr, "[+] "+format+"\n", a...)
	}
}

func Warn(format string, a ...any) {
	if UseColor {
		fmt.Fprintf(os.Stderr, Yellow+"[!] "+Reset+format+"\n", a...)
	} else {
		fmt.Fprintf(os.Stderr, "[!] "+format+"\n", a...)
	}
}

func Banner(name, version, owner string) {
	if Quiet {
		return
	}
	if !UseColor {
		fmt.Fprintf(os.Stderr, "== %s v%s by %s ==\n\n", name, version, owner)
		return
	}
	line := "══════════════════════════════════════════════════"
