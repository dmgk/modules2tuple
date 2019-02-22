modules2tuple is a helper tool for generating GH_TUPLE from vendor/modules.txt

#### Installation

    go get github.com/dmgk/modules2tuple

#### Usage

Vendor dependencies and run modules2tuple on vendor/modules.txt:

    go mod vendor
    modules2tuple vendor/modules.txt

By default, generated GH_TUPLE entries will place packages under `vendor`. This
can be changed by passing different prefix using -prefix option (e.g. `-prefix src`).
