package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	debug = false
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

//TODO: publish, CHANGE NAME!!!
//TODO: make a change to test_mac.sh and check that it works remotely (git stash; check; git stash pop)

func slurp(path string) string {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read %s: %s", path, err)
		os.Exit(1)
	}
	return string(buf)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "wrong number of arguments\nusage: newlint <before> <after> <source-diff>\n")

		os.Exit(1)
	}

	beforePath := os.Args[1]
	afterPath := os.Args[2]
	sourceDiff := os.Args[3]

	linterOutSecond := parseLinterOut(slurp(afterPath))

	da, err := parseDiff(slurp(sourceDiff))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	linterOutFirst := parseLinterOut(slurp(beforePath))

	if debug {
		fmt.Printf("Merge base has %d linter lines in files touched by diff\n", len(linterOutFirst))
	}

	mapToRight(linterOutFirst, da)

	linterOutFirstMap := make(map[pos]bool)
	for i := range linterOutFirst {
		linterOutFirstMap[linterOutFirst[i].pos] = true
	}

	bad := false
	for i := range linterOutSecond {
		if !linterOutFirstMap[linterOutSecond[i].pos] {
			fmt.Printf("%s:%d:%s\n", linterOutSecond[i].path, linterOutSecond[i].lineno, linterOutSecond[i].remark)
			bad = true
		}
	}

	if bad {
		os.Exit(1)
	}
}

func mapToRight(linterOut []linterError, da *diffAlignment) {
	for i := range linterOut {
		le := &linterOut[i]
		fa := da.leftToRight[le.path]
		if fa == nil {
			continue
		}
		if debug {
			fmt.Printf("%s -> %s\n", le.path, fa.toPath)
		}
		le.path = fa.toPath
		for j := len(fa.lines) - 1; j >= 0; j-- {
			if le.lineno >= fa.lines[j][0] {
				delta := fa.lines[j][1] - fa.lines[j][0]
				if debug {
					fmt.Printf("\t%d -> %d\n", le.lineno, le.lineno+delta)
				}
				le.lineno += delta
				break
			}
		}
	}
}
