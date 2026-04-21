# GEP-4359: Gateway API Regex

* Issue: [#4359](https://github.com/kubernetes-sigs/gateway-api/issues/4359)
* Status: Provisional

## TLDR

Regular expressions provide a powerful and concise method for traffic routing and manipulation, but present portability challenges due to the variety of regex engines.
The goal of this proposal is to define a common regex syntax and semantics that all implementations should support.
We use POSIX ERE (as defined in Chapter 9 of the [The Open Group Base Specifications Issue 8]) as the common denominator, with some feature exclusions.

## Goals

* Define a common regex syntax and semantics (Gateway API Regex) for Gateway API that all implementations should support.
* Define the set of inputs for our regex expressions.
* Define methods of referencing capturing groups.

## Non-Goals

* Limit regex features to a common subset.
* Methods of validating regular expressions at the API level
* Define specific use cases for regular expressions.
* Define failure modes when a regular expression or input is unsupported. Such validation requires a context-free grammar, and we cannot do that with CEL.

## Introduction/Overview

Regular expressions are a concise mechanism to describe large sets of strings, and for purposes of traffic routing and manipulation.
Unfortunately, the many regular expression engines ([RE2](https://github.com/google/re2), [PCRE](https://www.pcre.org/), and [rust-lang/regex](https://docs.rs/regex/latest/regex/))
present a portability challenge for Gateway API, as different proxies use different engines which support different syntax and features.

## Purpose

Regular expressions allow users to concisely express patterns for traffic routing and manipulation.
Although out of scope for this proposal, some use cases are:

* Path/URL matching
* URL rewriting
* Header matching

The goal here is to define a common regex syntax and semantics that all implementations should support.
Individual implementations can support more features, but having a common denominator allows for portable and testable behavior.
For example, a **hypothetical** path rewrite filter could look something like

```yaml
pathRewrite:
  pattern: ^/foo/(.*)$
  replacement: /bar/\1
```

Since the above pattern and replacement only use features supported by Gateway API Regex, its behavior should be consistent across implementations.

Individual implementations can support more extended regex features, such as RE2:

```yaml
pathRewrite:
  pattern: ^/foo/(?P<name>.*)$
  replacement: /bar/${name}
```

but any portability guarantees are now lost.

In other words, any pattern (and replacement) will implement Gateway API Regex syntax and semantics, but implementations can support additional syntax and semantics as well.

## Implementation and Support

The most popular Regex engines are RE2 and PCRE.

| Proxy          | Engine         |
|----------------|----------------|
| Envoy          | RE2            |
| HAProxy        | PCRE           |
| NGINX          | PCRE           |
| Traefik        | RE2            |


## Inputs

For Gateway API Regex we only explicitly support strings composed of ASCII characters that are not null, not control characters, and not spaces (other than space and tab) as inputs.
In other words, only the following code points are supported:

* `0x09` (tab)
* `0x20` (space)
* `0x21-0x7E` (printable characters)

The reason for this restriction is that regex engines differ in their support for line breaks, unicode, and control characters.
Supporting such inputs would add ambiguity and diminish portability with little practical benefits as such characters are rarely used in traffic routing and manipulation 
(notably, [HTTP header values can have `0x80-0xFF` code points](https://datatracker.ietf.org/doc/html/rfc9110#section-5.5), but their use is rare).
For example, in RE2, the `.` character might or might not match line breaks, depending on configuration.
`rust-lang/regex` treats multi-byte unicode characters as a single character, while in RE2, they are treated as a single character.
Null and control characters should not show up in traffic routing.

Some examples of valid input strings are

* `foo` (`0x66 0x6F 0x6F`)
* ` foo bar` (`0x20 0x66 0x6F 0x6F 0x20 0x62 0x61 0x72`)
* `foo\tbar` (`0x66 0x6F 0x6F 0x09 0x62 0x61 0x72`)

Some examples of invalid input strings are (implementations need not reject these, but they are not explicitly supported by the spec):

* `foo\nbar` (`0x66 0x6F 0x6F 0x0A 0x62 0x61 0x72`)
* `ｆ๏๏ ß∆я` (`0xef 0xbd 0x86 0xe0 0xb9 0x8f 0xe0 0xb9 0x8f 0xe2 0x80 0x82 0xc3 0x9f 0xe2 0x88 0x86 0xd1 0x8f`)


## Gateway API Regex Definition

We use IEEE POSIX ERE (as defined in Chapter 9 of the [The Open Group Base Specifications Issue 8]), as the base of Gateway API Regex, with a few exceptions:

* Matches are leftmost-first rather than leftmost-longest (quantifiers are still greedy).
* Unescaped square brackets in character classes are undefined (`[[]]`, `[]]`, `[[]`, `[[:alpha:]`, `[[=a=]]`, `[[.ch.]]`).
* Backslashes (`\`) retain their special meaning in character classes and can escape special characters, even if those special characters lose their special meaning in character classes. For example, `[\-]` matches `-`, and `[\]]` matches `]`.
* Consecutive quantifiers are undefined (`a**`, `a*?`, `a{1,2}?`).
* Any locale-specific behavior will assume the [C/POSIX locale](https://pubs.opengroup.org/onlinepubs/7908799/xbd/locale.html) (e.g. character ordering).

IEEE POSIX ERE is a good common denominator because
* The set of supported features is small (e.g. no backreferences)
* Broadly compatible across regex engines (RE2, rust-lang/regex, PCRE), especially because we don't have to worry about unicode, line breaks, and control sequences.

Some notable omissions from Gateway API Regex are:
* Backreferences (e.g. `\1`, `\2`, etc.)
* Matching flags (e.g. `(?i)` for case-insensitive matching)
* `\d` for digit, `\w` for word character, `\s` for whitespace character, and escaping any character that isn't a special character.
* Hex encoding of characters (e.g. `\xFF` or `\uFFFF`)
* Lazy vs non-lazy matching (e.g. `*` vs `*?`)
* Interval range expressions without starting numbers (`a{,2}`, `a{,}`).

### Examples

For the pattern `^(https?://)?www\.example\.(com|org)$`, the following strings would contain a match

* `www.example.com`
* `www.example.org`
* `http://www.example.com`
* `https://www.example.com`
* `http://www.example.org`
* `https://www.example.org`

The following strings would not match

* `example.com`
* `www.example.net`
* `www.example.(com|org)`
* `www.example.corg`
* `www.example.comorg`
* `ahttp://www.example.org`
* `http://www.example.orga`

For the pattern `a*`, the following strings would match

* ``
* `a` (many possible matches)
* `aaaaa` (many possible matches)
* `b` (many possible matches)
* In fact, any string would contain a match

For the pattern `a+`, the following strings would match

* `a`
* `baaaaab` (many possible matches).

The following strings would not match

* ``
* `b`

The patterns `^colo(u|)r$`, `^colou?r$`, `^colo(u)?r$`, `^colo(u)?r$`, `^colou{0,1}r$` are all equivalent, and would match

* `color`
* `colour`

The following strings would not match

* `colouur`
* `colo(u)r`

The pattern `\*+` would match

* `*`
* `****`
* `\*+`

The pattern `[a-zA-Z0-9.]` would match

* `a`
* `Z`
* `5`
* `.`

The following strings would not match

* `-`
* ` `

The pattern `[^a-zA-Z]` would match

* `-`
* ` `
* `9`

The following strings would not match

* `a`
* `Z`
* `k`

For the pattern `^[^^]$`, the only non-matching strings are

* ``
* `^`
* Any string with two or more characters

For the patterns `^[-a]$` and `^[a-]$`, the only matching strings are 

* `-`
* `a`

For the equivalent patterns `^[-.*(){}|^$\\]$` and `^[\-\.\*\(\)\{\}\|\^\$\\]$`, the only matching strings are

* `-`
* `.`
* `*`
* `(`
* `)`
* `{`
* `}`
* `|`
* `^`
* `$`
* `\`

For the pattern `^[\-Z]$`, the only matching strings are

* `-`
* `Z`

## Replacement

Replacement is usually straightforward, but has some semantic ambiguities.
Given a pattern, a replacement, and an input, we allow higher level APIs to define semantics such as
* whether the first or all matches should be replaced.
* if replacing all matches, repeatedly find the next non-overlapping leftmost-first match, replace it, then continue scanning after the replaced match.
    * This means that some instances of patterns might be skipped if there are overlaps.
    * The resulting string might contain new matches due to replacements.

### Referencing Capturing Groups

In the case that referencing capturing groups is supported, APIs MUST use the `\1`, `\2`, etc. syntax in the replacement string to reference capturing groups.
Named capturing groups are not explicitly supported, and the syntax for referencing them is undefined.
A capturing group is any expression enclosed in unescaped parentheses.
The order of capturing groups is determined by the order of the opening parentheses, from left to right in the pattern, starting at 1.
If the replacement references a non-existent capturing group, the reference is treated as an empty string.
If a capturing group has a quantifier at the end, the reference in the replacement string will be replaced with the longest match of that capturing group.
Here are some examples when replacing all instances of the pattern with the replacement:

| Pattern         | Replacement | Input           | Output          |
|-----------------|-------------|-----------------|-----------------|
| `aba`           | `c`         | `abaaba`        | `cc`            |
| `aba`           | `c`         | `ababa`         | `cba`           |
| `aba`           | `a`         | `ababa`         | `aba`           |
| `a([bc])+a`     | `\1`        | `abca`          | `c`             |
| `(a\|ab)`       | `c`         | `ab`            | `cb`            |

Note in the last example, the pattern is `(a|ab)`, the backslash is to escape the `|` in the markdown table.
The output is `cb` rather than `c` because of leftmost-first matching.

