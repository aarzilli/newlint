package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type diffAlignment struct {
	leftToRight map[string]*fileAlignment
}

type fileAlignment struct {
	toPath string
	lines  [][2]int
}

type lineScanner struct {
	*bufio.Scanner
	lineno int
}

func (s *lineScanner) Scan() bool {
	s.lineno++
	return s.Scanner.Scan()
}

func (s *lineScanner) error(reason string) error {
	return fmt.Errorf("%d: %s.\n\tdiff line %q", s.lineno, reason, s.Text())

}

func parseDiff(in string) (da *diffAlignment, err error) {
	defer func() {
		ierr := recover()
		if ierr == nil {
			return
		}
		err2, ok := ierr.(error)
		if ok {
			err = err2
		} else {
			panic(ierr)
		}
	}()

	const diffPrefix = "diff --git "

	da = &diffAlignment{}
	var fa *fileAlignment

	da.leftToRight = make(map[string]*fileAlignment)

	s := &lineScanner{bufio.NewScanner(bytes.NewBuffer([]byte(in))), 0}
	for s.Scan() {
		line := s.Text()
		switch {
		case strings.HasPrefix(line, diffPrefix):
			leftPath, rightPath := parseDiffHeader(s)
			fa = &fileAlignment{toPath: rightPath}
			da.leftToRight[leftPath] = fa

		case len(line) > 0 && line[0] == '@':
			if fa == nil {
				panic(s.error("chunk outside of a diff section"))
			}
			leftLineno, leftSize, rightLineno, rightSize, _ := parseChunkHeader(s, line)
			fa.lines = append(fa.lines, [2]int{leftLineno + leftSize, rightLineno + rightSize})

		case len(line) > 0 && (line[0] == ' ' || line[0] == '-' || line[0] == '+'):
			// ignore
		default:
			panic(s.error("unexpected line"))
		}
	}
	must(s.Err())
	return da, nil
}

func parseDiffHeader(s *lineScanner) (leftPath, rightPath string) {
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "--- ") {
			leftPath = line[6:]
			break
		}
	}

	if leftPath == "" {
		panic(s.error("could not find from/to header while trying to parse a diff section"))
	}

	if !s.Scan() {
		panic(s.error("could not find from/to header while trying to parse a diff section"))
	}
	if !strings.HasPrefix(s.Text(), "+++ ") {
		panic(s.error("expected second line of a from/to header"))
	}

	rightPath = s.Text()[6:]

	return leftPath, rightPath
}

func parseChunkHeader(s *lineScanner, line string) (leftLineno, leftSize, rightLineno, rightSize int, contents string) {
	v := strings.Split(line, "@@")
	ranges := strings.Split(strings.TrimSpace(v[1]), " ")
	var err error

	leftLineno, err = strconv.Atoi(strings.Split(ranges[0], ",")[0])
	if err != nil {
		panic(s.error("malformed chunk header"))
	}
	leftLineno *= -1
	rightSize, err = strconv.Atoi(strings.Split(ranges[0], ",")[1])
	if err != nil {
		panic(s.error("malformed chunk header"))
	}

	rightLineno, err = strconv.Atoi(strings.Split(ranges[1], ",")[0])
	if err != nil {
		panic(s.error("malformed chunk header"))
	}
	rightSize, err = strconv.Atoi(strings.Split(ranges[1], ",")[1])
	if err != nil {
		panic(s.error("malformed chunk header"))
	}

	contents = v[2][1:]
	return
}
