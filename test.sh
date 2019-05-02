#!/bin/bash

mkdir tmp

test() {
    expected=$1
    input=$2
    echo "int main() { $input }" > ./tmp/tmp.c
    ./mycc ./tmp/tmp.c > tmp/tmp.s
    gcc -o ./tmp/tmp tmp/tmp.s tmp/hoge.o
    ./tmp/tmp
    actual="$?"

    if [[ "$actual" = "$expected" ]]; then
        echo "$input => $expected"
    else
        echo "$expected expected, but got $actual"
        exit 1
    fi
}

#test 3 "a = 3;"
#test 8 "a = 1 + 3 + 4;"
#test 7 "a = 1 + 10 - 4;"
#test 14 "a = 1 * 2 + 3 * 4;"
#test 11 "a = 1 + 2 * 3 + 4;"
#test 20 "a = (1 * 2 + 3) * 4;"
#test 21 "a = (1 + 2) * (3 + 4);"
#test 3 "a = 1; b = 3;"
#test 1 "a = 1; b = 3; return 1;"
#test 23 "a = 3 + 10 * 2; b = 3; return a;"
#test 3 "a = 3 + 10 * 2; b = 3; return b;"
#test 8 "a = 10 + 1 * -2;"
#test 12 "a = 10 - 1 * -2;"

#test 1 "int a = 1 == 1; return a;"
#test 0 "int a = 1 == 2; return a;"
#test 3 "return foo(1, 2, 3, 4, 5, 6);"
#test 10 "debug2(10); return 10;"
#test 1 "return bar(1); } int bar(){ return 1;"
#test 1 "return bar(1); } int bar(int a){ return 1;"
#test 16 "return bar(3, 5); } int bar(int a, int b){ return 1 + a * b;"
#test 10 "if (1 == 1) { return 10; } return 11;"
#test 11 "if (1 == 2) { return 10; } return 11;"
#test 10 "if (1 == 1) return 10; return 11;"
#test 11 "if (1 == 2) return 10; return 11;"
#test 3 "int i = 0; while (i != 3) { i = i + 1; } return i;"
#test 1 "int i = 0; while (i != 3) { i = i + 1; debug(i); } return 1;"
#test 3 "int i = 0; while (i != 3) { i = i + 1; int j = debug(i); } return i;"
#test 10 "for (int i = 0; i != 10; i = i + 1) { debug(i); } return i;"
#test 10 "for (int i = 0; i != 10; i = i + 1) debug(i); return i;"
#test 10 "int x; x = 10; int *y; y = &x; return *y;"
#test 4 "int *p; alloc4(&p, 1, 2, 4, 8); int *q; q = p + 2; return *q;"
#test 8 "int *p; alloc4(&p, 1, 2, 4, 8); int *q; q = p + 3; return *q;"
#test 4 "return sizeof(4);"
#test 4 "int i = 10; return sizeof(i);"
#test 4 "int i = 10; return sizeof(i + 4);"
#test 8 "int *i; return sizeof(i);"
#test 2 "int i[10]; int b = 2; return b;"
#test 10 "int i[10]; i[0] = 10; return i[0];"
#test 11 "int i[10]; i[1] = 11; return i[1];"
test 1 "int a[2]; *a = 1; return a[0];"
test 2 "int a[2]; *(a + 1) = 2; return a[1];"
test 3 "int a[2]; *a = 1; *(a + 1) = 2; int *p; p = a; return *p + *(p + 1);"
echo OK
