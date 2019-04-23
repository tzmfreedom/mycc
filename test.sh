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

test 0 0
test 42 42
echo OK