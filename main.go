package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
)

const (
	TMPPATH = "/tmp"
	REPO    = "https://github.com/go-delve/delve"
	//EXAMPLE_PULL_REQUEST = 2097
	EXAMPLE_PULL_REQUEST = 2011
	debug                = true
)

var LINTER_COMMAND = []string{"golangci-lint", "run", "--max-issues-per-linter", "0", "--max-same-issues", "0", "--print-issued-lines", "false", "--uniq-by-line", "false", "--timeout", "10m"}
var repodir string

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
	repodir = filepath.Join(TMPPATH, filepath.Base(REPO))
	reponame := filepath.Base(REPO)
	repowner := strings.Split(REPO, "/")[3]

	if _, err := os.Stat(repodir); os.IsNotExist(err) {
		if debug {
			fmt.Printf("cloning\n")
		}
		execIn(TMPPATH, "git", "clone", REPO+".git")
	} else {
		if debug {
			fmt.Printf("updating\n")
		}
		execIn(repodir, "git", "checkout", "master")
		//execIn(repodir, "git", "pull", "origin")
	}

	client := github.NewClient(nil)
	pr, _, err := client.PullRequests.Get(context.Background(), repowner, reponame, EXAMPLE_PULL_REQUEST)
	must(err)

	headlabel := strings.Split(*pr.Head.Label, ":")
	pruser := headlabel[0]
	prbranch := headlabel[1]

	if debug {
		fmt.Printf("fetching %q %q\n", pruser, prbranch)
	}
	execIn(repodir, "git", "fetch", fmt.Sprintf("https://github.com/%s/%s.git", pruser, reponame), prbranch)
	mergeBase := strings.TrimSpace(execIn(repodir, "git", "merge-base", "FETCH_HEAD", "master"))

	if debug {
		fmt.Printf("merge base is %q\n", mergeBase)
	}

	da, err := parseDiff(execIn(repodir, "git", "diff", "--no-color", mergeBase, "FETCH_HEAD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	execIn(repodir, "git", "checkout", mergeBase)
	linterOutMergeBaseStr := execInNoErr(repodir, LINTER_COMMAND[0], LINTER_COMMAND[1:]...)
	linterOutMergeBase := parseLinterOut(linterOutMergeBaseStr, da, true)

	if debug {
		fmt.Printf("Merge base has %d linter lines in files touched by diff\n", len(linterOutMergeBase))
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

	//TODO:
	// - parse diff and align source
	// - parse linter output and align

}

func repoAbsPath(p string) string {
	if p == "" {
		return ""
	}
	if p[0] == '/' {
		//TODO: doesn't work on windows
		return p
	}
	return filepath.Join(repodir, p)
}
