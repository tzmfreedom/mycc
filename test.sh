#!/bin/bash

test() {
    expected=$1
    input=$2
    ./mycc "$input" > tmp.s
    gcc -o tmp tmp.s
    ./tmp
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
echo OK