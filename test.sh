#!/bin/bash

mkdir tmp

test() {
    expected=$1
    input=$2
    ./mycc "$input" > tmp/tmp.s
    gcc -o ./tmp/tmp tmp/tmp.s
    ./tmp/tmp
    actual="$?"

    if [[ "$actual" = "$expected" ]]; then
        echo "$input => $expected"
    else
        echo "$expected expected, but got $actual"
        exit 1
    fi
}

test 3 "a = 3;"
test 8 "a = 1 + 3 + 4;"
test 7 "a = 1 + 10 - 4;"
test 14 "a = 1 * 2 + 3 * 4;"
test 11 "a = 1 + 2 * 3 + 4;"
test 20 "a = (1 * 2 + 3) * 4;"
test 21 "a = (1 + 2) * (3 + 4);"
test 3 "a = 1; b = 3;"
test 1 "a = 1; b = 3; return 1;"
test 23 "a = 3 + 10 * 2; b = 3; return a;"
test 3 "a = 3 + 10 * 2; b = 3; return b;"
echo OK
