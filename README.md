## modules2tuple

Helper tool for generating GH_TUPLE and GL_TUPLE from vendor/modules.txt.

![Tests](https://github.com/dmgk/modules2tuple/actions/workflows/tests.yml/badge.svg)

#### Installation

As a FreeBSD binary package:

    pkg install modules2tuple

or from ports:

    make -C /usr/ports/devel/modules2tuple install clean

To install latest dev version directly from GitHub:

    go install github.com/dmgk/modules2tuple/v2

#### Usage

    modules2tuple [options] modules.txt

    Options:
        -offline  disable all network access (env M2T_OFFLINE, default false)
        -debug    print debug info (env M2T_DEBUG, default false)
        -v        show version

    Usage:
        Vendor package dependencies and then run modules2tuple with vendor/modules.txt:

        $ go mod vendor
        $ modules2tuple vendor/modules.txt

    When running in offline mode:
        - mirrors are looked up using static list and some may not be resolved
        - milti-module repos and version suffixes ("/v2") are not automatically handled
        - Github tags for modules ("v1.2.3" vs "api/v1.2.3") are not automatically resolved
        - Gitlab commit IDs are not resolved to the full 40-char IDs
