package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	debug = true
)

var LinterCommand = []string{"staticcheck", "-checks", "all,-ST1003"}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func execIn(dir string, cmdstr string, args ...string) string {
	if debug {
		fmt.Printf("\t%s: %q %q\n", dir, cmdstr, args)
	}
	cmd := exec.Command(cmdstr, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error %v executing %q %q: %s\n", err, cmdstr, args, string(out))
		os.Exit(1)
	}
	return string(out)
}

func execInNoErr(dir string, cmdstr string, args ...string) string {
	if debug {
		fmt.Printf("\t%s: %q %q\n", dir, cmdstr, args)
	}
	cmd := exec.Command(cmdstr, args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out)
}

//TODO: change to taking three files:
// 	- first output
// 	- second output
// 	- git diff
//TODO: check that this works locally
//TODO: publish, CHANGE NAME!!!
//TODO: make a change to test_mac.sh and check that it works remotely (git stash; check; git stash pop)

func main() {
	firstCommit := "master"
	
	linterOutSecond := parseLinterOut(execInNoErr(".", LinterCommand[0], LinterCommand[1:]...))

	da, err := parseDiff(execIn(".", "git", "diff", "--no-color", firstCommit, execIn(".", "git", "rev-parse", "HEAD")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	LinterCommand = append(LinterCommand, strings.TrimSpace(execIn(".", "go", "list", "-m"))+"/...")

	execIn(".", "git", "checkout", firstCommit)
	linterOutFirst := parseLinterOut(execInNoErr(".", LinterCommand[0], LinterCommand[1:]...))

	if debug {
		fmt.Printf("Merge base has %d linter lines in files touched by diff\n", len(linterOutFirst))
	}
	
	mapToRight(linterOutFirst, da)
	
	execIn(".", "git", "checkout", "@{-1}")

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
