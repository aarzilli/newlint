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

func main() {
	firstCommit := "HEAD^"
	secondCommit := "HEAD"

	da, err := parseDiff(execIn(".", "git", "diff", "--no-color", firstCommit, secondCommit))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	LinterCommand = append(LinterCommand, strings.TrimSpace(execIn(".", "go", "list", "-m"))+"/...")

	execIn(".", "git", "checkout", firstCommit)
	linterOutFirst := parseLinterOut(execInNoErr(".", LinterCommand[0], LinterCommand[1:]...))

	if debug {
		fmt.Printf("Merge base has %d linter lines in files touched by diff\n", len(linterOutFirst))
	}

	/*execIn(repodir, "git", "checkout", "FETCH_HEAD")
	linterOutFetchHeadStr := execInNoErr(repodir, LINTER_COMMAND[0], LINTER_COMMAND[1:]...)
	linterOutFetchHead := parseLinterOut(linterOutFetchHeadStr, da, false)
	if debug {
		fmt.Printf("FETCH_HEAD has %d linter lines in files touched by diff\n", len(linterOutFetchHead))
		for _, e := range linterOutFetchHead {
			fmt.Printf("%s:%d: %s\n", e.path, e.lineno, e.remark)
		}
	}*/

	_ = da

	//TODO:
	// - parse diff and align source
	// - parse linter output and align

}
