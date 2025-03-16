# Naming Conventions

This file specifies the naming conventions for identifiers in this library.  
The convention applies to all identifiers except packages,
including variables, constants, types (including everything declared with
the keyword `type`, such as structures, interfaces, type aliases, ...),
functions, methods, parameters (including function parameters and
type parameters), function results, method receivers, and labels.

## Notice

In this convention, by default, the "lowercase" and "uppercase" of
a word, phrase, name, identifier, or abbreviation indicate that
its first letter, rather than all letters, is lowercase/uppercase.  
For example, `numCPU` is lowercase and `NumCPU` is uppercase.

## Abbreviation

Abbreviations should be used to make code concise
without increasing ambiguity or difficulty in understanding.  
In general, the smaller the scope (block) of an identifier,
the shorter the abbreviation it should use.  
Except for structure fields and interface methods,
identifiers with a top-level scope (package block)
should not use any abbreviations that are not widely known.

Different words/phrases may have the same abbreviation.  
Developers are responsible for ensuring that
abbreviations used are unambiguous in their context.

This section lists the abbreviations used in this library.  
Some well-known abbreviations, such as `API`, `CPU`, `ID`, and `IP`,
are not listed.

The abbreviations are listed in alphabetical order.
The order does not reflect usage frequency or recommendation level.

The abbreviations in the table below are given in lowercase.  
By default, the uppercase capitalizes only the first letter.  
For example, `buf` &rarr; `Buf`, `num` &rarr; `Num`.  
If the abbreviation has special capitalization,
it should be given in parentheses.  
For example, `inout (InOut)` represents an abbreviation
whose lowercase is `inout` and uppercase is `InOut`.

Symbols used in the table below:

- `x-`: x is a prefix.
- `*x`: x is rarely used, or only to avoid colliding with
keywords or other identifiers such as standard packages.
- `^x`: x is a singleton, not part of others.

| Word or Phrase                          | Abbreviations       |
| --------------------------------------- | ------------------- |
| buffer                                  | ^b, buf, ^p         |
| builder                                 | ^b                  |
| continue                                | \*cont              |
| counter                                 | ctr                 |
| default                                 | \*dflt              |
| function                                | ^f, fn, func        |
| input/output                            | \*inout (InOut), io |
| iterator                                | iter                |
| number (indicating amount, quantity)    | n\-, num\-          |
| number (indicating rank, serial number) | no                  |
| pointer                                 | p\-, ptr            |
| type                                    | typ                 |
| value                                   | ^v, val             |

This table may be extended in the future.

Abbreviations that can be used are not limited to this table.
However, for new abbreviations, some comments or documents should be attached.

## Boolean

The name of a Boolean value (including variables, constants, function
parameters, function results, and functions and methods that return a Boolean)
should indicate the state or the action to be taken when the value is true.

Unlike some common naming conventions,
the name does not need to be in the form of a general question.  
For example, use `valid` instead of `isValid`;
use `skip` instead of `doesSkip`.  
In particular, if the name causes ambiguity or collides with others,
or if its general question form is widely used in other libraries
(e.g., `isSorted`), then the general question form should be used instead.

## Capitalization

The lowercase of identifiers follows the "camelCase" (lower camel case)
and the uppercase follows "PascalCase" (upper camel case).

In particular, common abbreviations, such as `API`, `CPU`, `ID`, and `IP`,
should be capitalized the same way they are normally written.  
For example, use `id` (lowercase) and `ID` (uppercase) instead of `Id`;
use `userIP` (lowercase) and `UserIP` (uppercase)
instead of `userIp` or `UserIp`;
use `numCPU` (lowercase) and `NumCPU` (uppercase)
instead of `numCpu` or `NumCpu`.

## Iterator

As of Go 1.23, Go has introduced support for iterators.
For details see the [`iter` package documentation](https://go.dev/pkg/iter "iter package"),
the [language specification](https://go.dev/ref/spec#For_range "The Go Programming Language Specification - For statements with range clause"),
and the [Range over Function Types blog post](https://go.dev/blog/range-functions "Range Over Function Types").

The [Go official documentation](https://pkg.go.dev/iter#hdr-Naming_Conventions "iter package - Naming Conventions")
specifies the naming conventions for iterator functions and methods.  
However, this file requires that in addition to the official naming conventions,
iterator functions and methods must use the prefix `iter-`
to distinguish them from the functions and methods that return
collections such as `[]T` and `map[K]V`.  
For example, the iterator method should be named as follows:

```go
// IterAll returns an iterator over all elements in s.
func (s *Set[V]) IterAll() iter.Seq[V]
```

instead of that shown in the official documentation:

```go
// All returns an iterator over all elements in s.
func (s *Set[V]) All() iter.Seq[V]
```

So that it can be distinguished from a possibly existing list method:

```go
// All returns a list of all elements in s.
func (s *Set[V]) All() []V
```

## "*The Number Of*" and "*The Pointer Of*"

"*The number of*" should use the prefix `num-`,
followed by the singular form (e.g., `numItem`).

"*The pointer of*" should use the suffix `-ptr` (e.g., `itemPtr`).

For identifiers with a function-level or smaller scope
(inside a function block):  
"*The number of*" can also use the prefix `n-`,
followed by the plural form (e.g., `nItems`).  
"*The pointer of*" can also use the prefix `p-` (e.g., `pItem`).
