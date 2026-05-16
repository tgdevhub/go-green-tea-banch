#!/usr/bin/env bash

set -euo pipefail

COUNT="${COUNT:-5}"
BENCHTIME="${BENCHTIME:-1s}"
OUT_DIR="${OUT_DIR:-./bench-results}"

mkdir -p "$OUT_DIR"

echo ">> Стандартный GC (count=$COUNT, benchtime=$BENCHTIME)"
go test -run='^$' -bench=. -benchmem \
    -count="$COUNT" -benchtime="$BENCHTIME" \
    | tee "$OUT_DIR/standard.txt"

echo
echo ">> Green Tea GC (count=$COUNT, benchtime=$BENCHTIME)"
GOEXPERIMENT=greenteagc go test -run='^$' -bench=. -benchmem \
    -count="$COUNT" -benchtime="$BENCHTIME" \
    | tee "$OUT_DIR/greentea.txt"

echo
if command -v benchstat >/dev/null 2>&1; then
    echo ">> Сравнение через benchstat:"
    benchstat "$OUT_DIR/standard.txt" "$OUT_DIR/greentea.txt"
else
    echo "benchstat не найден. Установить:"
    echo "  go install golang.org/x/perf/cmd/benchstat@latest"
fi
