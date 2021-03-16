#!/bin/bash

C=all,-ST1003
P=$(go list -m)/...
staticcheck -checks $C $P > /tmp/linter-after
git diff --no-color master HEAD > /tmp/source-diff
git checkout master
staticcheck -checks $C $P > /tmp/linter-before
newlint /tmp/linter-before /tmp/linter-after /tmp/source-diff
git checkout @{-1}
