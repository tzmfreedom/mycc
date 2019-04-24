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

test 3 "3"
test 8 "1 + 3 + 4"
test 7 "1 + 10 - 4"
test 14 "1 * 2 + 3 * 4"
test 11 "1 + 2 * 3 + 4"
echo OK
