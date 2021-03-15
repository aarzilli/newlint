package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type linterError struct {
	path   string
	lineno int
	remark string
}

var reFileCol = regexp.MustCompile(`^([0-9A-Za-z_/.]+):(\d+)(?::\d+)?: (.*)$`) //TODO: doesn't work with paths containing spaces

func parseLinterOut(in string) []linterError {
	lines := strings.Split(in, "\n")
	count := 0
	r := make([]linterError, 0, len(lines))
	for _, line := range lines {
		m := reFileCol.FindStringSubmatch(line)
		if len(m) != 4 {
			continue
		}

		count++

		path := m[1]
		lineno, _ := strconv.Atoi(m[2])
		remark := m[3]

		r = append(r, linterError{path, lineno, remark})
	}
	if debug {
		fmt.Printf("recognized linter output %d/%d\n", count, len(lines))
	}
	sort.Slice(r, func(i, j int) bool {
		if r[i].path == r[j].path {
			return r[i].lineno < r[j].lineno
		}
		return r[i].path < r[j].path
	})
	return r
}
