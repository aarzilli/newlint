package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type pos struct {
	path   string
	lineno int
}

type linterError struct {
	pos
	remark string
}

var reFileCol = regexp.MustCompile(`^([0-9A-Za-z_/.]+):(\d+):(.*)$`) //TODO: doesn't work with paths containing spaces

func parseLinterOut(in string) []linterError {
	lines := strings.Split(in, "\n")
	count := 0
	r := make([]linterError, 0, len(lines))
	for _, line := range lines {
		m := reFileCol.FindStringSubmatch(line)
		if len(m) != 4 {
			r[len(r)-1].remark += "\n" + line
			continue
		}

		count++

		path := m[1]
		lineno, _ := strconv.Atoi(m[2])
		remark := m[3]

		r = append(r, linterError{pos{path, lineno}, remark})
	}
	if debug {
		fmt.Printf("recognized linter output %d/%d\n", count, len(lines))
	}
	return r
}
