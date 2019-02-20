modules2tuple is a helper tool for generating GH_TUPLE from vendor/modules.txt

To install

    go get -u github.com/dmgk/modules2tuple

#### Usage

Vendor dependencies and run modules2tuple on vendor/modules.txt:

    go mod vendor
    modules2tuple vendor/modules.txt
