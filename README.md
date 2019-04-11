modules2tuple is a helper tool for generating GH_TUPLE from vendor/modules.txt

[![Build Status](https://travis-ci.org/dmgk/modules2tuple.svg?branch=master)](https://travis-ci.org/dmgk/modules2tuple)

#### Installation

As a FreeBSD binary package:

    pkg install modules2tuple

or from ports:

    make -C /usr/ports/devel/modules2tuple install clean

To install latest dev version directly from GitHub:

    go get github.com/dmgk/modules2tuple

#### Usage

Vendor dependencies and run modules2tuple on vendor/modules.txt:

    go mod vendor
    modules2tuple vendor/modules.txt

By default, generated GH_TUPLE entries will place packages under `vendor`. This
can be changed by passing different prefix using -prefix option (e.g. `-prefix src`).


#### Contributing

modules2tuple knows about some GitHub mirrors and can automatically generate correct
GH_TUPLE entries for them, but it always needs more. Please open a pull request or create an
issue if you find some package names that it cannot handle (it leaves them commented out).
