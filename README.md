Hermes is an experimental framework for building compilers in Go. It's
inspired by the [Nanopass Framework](https://nanopass.org/) and
[attribute grammars](https://en.wikipedia.org/wiki/Attribute_grammar),
particularly [UUAGC](https://github.com/UU-ComputerScience/uuagc/tree/master/uuagc/trunk/doc).

# Current status

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

# Future direction

The current Hermes prototype was a time-bounded experiment at porting
scheme-to-c to work out concrete details and discover implementation
hurdles. I don't plan to work on it further in the immediate future,
but I hope to revisit eventually.

Some immediate reflections:

Writing the language definitions and passes in Go syntax is convenient
because it benefits from existing Go support in IDEs. E.g., I
anticipate eventually that referencing standard Go types and functions
will be useful, and by writing directly in Go syntax allows code
navigation features to work. Certainly writing passes is made somewhat
easier because of editor support for referring to the generated
structs/interfaces for languages.

However, I'm not fully convinced yet that it's better than defining a
custom syntax, as tools like yacc have traditionally done. For
example, the https://github.com/a-h/templ project adds JSX-like syntax
to Go and introduces new ".templ" files to delineate this change in
syntax/semantics, and yet it still advertises good editor support by
proxing gopls.
