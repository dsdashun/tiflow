#!/bin/bash

mydir=$(dirname "${BASH_SOURCE[0]}")

grammar_dir="${mydir}/../grammar"
target_dir="${mydir}/../internal/parser"

pushd "${grammar_dir}"
java -Xmx500M -cp "$HOME/local/lib/antlr-4.9-complete.jar:$CLASSPATH" org.antlr.v4.Tool Workload.g4
popd

mv "${grammar_dir}/"*.go "${target_dir}/"
