#!/bin/bash

mydir=$(dirname "${BASH_SOURCE[0]}")

target_dir="${mydir}/../internal/parser"
grammar_dir_rel_to_target="../../grammar"

pkg_name=$(basename "${target_dir}")

cd "${target_dir}" #go to target dir, to ensure the comments of the generated Go file be reasonable

java -Xmx500M -cp "$HOME/local/lib/antlr-4.9-complete.jar:$CLASSPATH" org.antlr.v4.Tool -package "${pkg_name}" "${grammar_dir_rel_to_target}/Workload.g4"

for file_name in "${grammar_dir_rel_to_target}"/*.go; do
	base_file_name=$(basename "${file_name}")
	mv "${file_name}" "${base_file_name/.go/.antlr4.go}"
done
