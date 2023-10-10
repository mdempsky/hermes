Hermes is an experimental framework for building compilers in Go. It's
inspired by the [Nanopass Framework](https://nanopass.org/) and
[attribute grammars](https://en.wikipedia.org/wiki/Attribute_grammar),
particularly [UUAGC](https://github.com/UU-ComputerScience/uuagc/tree/master/uuagc/trunk/doc).

Currently, it's still very early development. Current artifacts:

* example/lang.go

This source file defines all of the sublanguages used through the
[scheme-to-c](https://github.com/akeep/scheme-to-c/blob/main/c.ss)
example nanopass compiler. It's written in Go-compatible syntax, but
the Go semantics are meaningless.

* cmd/mklang

This command, when run within the example subdirectory, transforms
lang.go into the lang/L* packages. One package per sublanguage.

* passes/*.go

These source files contain the first several passes of the scheme-to-c
compiler, translated into how I envision writing them in Hermes. These
are not actual Go source, but again Go-compatible syntax.

* cmd/passify

This command, when run within the example subdirectory, will
eventually turn the passes/*.go files into actual executable Go code.
