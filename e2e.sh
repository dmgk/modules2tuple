#!/bin/sh

set -e

for given in testdata/*_modules.txt; do
    m=${given#testdata/}
    m=${m%_modules.txt}
    expected=testdata/${m}_expected.txt
    actual=/tmp/${m}_actual.txt

    ./modules2tuple ${given} > ${actual}
    diff -u ${expected} ${actual}
    rm ${actual}
done
