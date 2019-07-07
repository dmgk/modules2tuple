modules2tuple is a helper tool for generating GH_TUPLE and GL_TUPLE from vendor/modules.txt

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

    $ go mod vendor
    $ modules2tuple vendor/modules.txt

By default, generated tuple entries will place packages under `vendor`. This
can be changed by passing different prefix using `-prefix` option (e.g. `-prefix src`).

When generating GL_TUPLE entries, modules2tuple will attempt to use Gitlab API to
resolve short commit IDs and tags to the full 40-character IDs as required by bsd.sites.mk. 
If network access is not available or not wanted, this commit ID translation can be disabled
with `-offline` flag.

#### Contributing

modules2tuple knows about some GitHub mirrors and can automatically generate correct
GH_TUPLE entries for them, but it always can use more. Please open a pull request or create an
issue if you find some package names that it cannot handle (it leaves them commented out).
