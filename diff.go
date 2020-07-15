package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type diffAlignment struct {
	leftToRight     map[string]map[int]int
	leftToRightPath map[string]string
	knownLeftLines  map[string]map[int]string
	knownRightLines map[string]map[int]string
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

	curpath := ""
	var rightLineno, leftLineno = -1, -1

	da = &diffAlignment{}

	da.leftToRight = make(map[string]map[int]int)
	da.leftToRightPath = make(map[string]string)
	da.knownLeftLines = make(map[string]map[int]string)
	da.knownRightLines = make(map[string]map[int]string)

	s := &lineScanner{bufio.NewScanner(bytes.NewBuffer([]byte(in))), 0}
	for s.Scan() {
		line := s.Text()
		switch {
		case strings.HasPrefix(line, diffPrefix):
			var rightPath string
			curpath, rightPath = parseDiffHeader(s)
			curpath = repoAbsPath(curpath)
			rightPath = repoAbsPath(rightPath)
			da.leftToRightPath[curpath] = rightPath
			if da.leftToRight[curpath] == nil {
				da.leftToRight[curpath] = make(map[int]int)
			}
			curpath = rightPath
			if da.knownLeftLines[curpath] == nil {
				da.knownLeftLines[curpath] = make(map[int]string)
			}
			if da.knownRightLines[curpath] == nil {
				da.knownRightLines[curpath] = make(map[int]string)
			}
			rightLineno, leftLineno = -1, -1

		case len(line) > 0 && line[0] == '@':
			if curpath == "" {
				panic(s.error("chunk outside of a diff section"))
			}
			contents := parseChunkHeader(s, line, &leftLineno, &rightLineno)
			da.knownLeftLines[curpath][leftLineno] = contents
			da.knownRightLines[curpath][rightLineno] = contents

		case len(line) > 0 && (line[0] == ' ' || line[0] == '-' || line[0] == '+'):
			if rightLineno < 0 || leftLineno < 0 {
				panic(s.error("diff line outside of a chunk"))
			}

			contents := line[1:]

			switch line[0] {
			case ' ':
				da.leftToRight[curpath][leftLineno] = rightLineno
				da.knownLeftLines[curpath][leftLineno] = contents
				da.knownRightLines[curpath][rightLineno] = contents
				leftLineno++
				rightLineno++
			case '-':
				da.leftToRight[curpath][leftLineno] = -1
				da.knownLeftLines[curpath][leftLineno] = contents
				leftLineno++
			case '+':
				da.knownRightLines[curpath][rightLineno] = contents
				rightLineno++
			}
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

func parseChunkHeader(s *lineScanner, line string, leftLineno, rightLineno *int) (contents string) {
	v := strings.Split(line, "@@")
	ranges := strings.Split(strings.TrimSpace(v[1]), " ")
	var err error
	*leftLineno, err = strconv.Atoi(strings.Split(ranges[0], ",")[0])
	if err != nil {
		panic(s.error("malformed chunk header"))
	}
	*leftLineno *= -1
	*rightLineno, err = strconv.Atoi(strings.Split(ranges[1], ",")[0])
	if err != nil {
		panic(s.error("malformed chunk header"))
	}
	return v[2][1:]
}
