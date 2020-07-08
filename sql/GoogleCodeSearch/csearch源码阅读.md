search源码阅读.md

## Csearch

## regexp.synstax
``` go
// A Regexp is a node in a regular expression syntax tree.
type Regexp struct {
    Op       Op // operator
    Flags    Flags
    Sub      []*Regexp  // subexpressions, if any
    Sub0     [1]*Regexp // storage for short Sub
    Rune     []rune     // matched runes, for OpLiteral, OpCharClass
    Rune0    [2]rune    // storage for short Rune
    Min, Max int        // min, max for OpRepeat
    Cap      int        // capturing index, for OpCapture
    Name     string     // capturing name, for OpCapture
}
```


## codesearch.regexp.Regexp
```go
// Regexp is the representation of a compiled regular expression.
// A Regexp is NOT SAFE for concurrent use by multiple goroutines.
type Regexp struct {
    Syntax *syntax.Regexp
    expr   string // original expression
    m      matcher
}

// A matcher holds the state for running regular expression search.
type matcher struct {
    prog      *syntax.Prog       // compiled program
    dstate    map[string]*dstate // dstate cache
    start     *dstate            // start state
    startLine *dstate            // start state for beginning of line
    z1, z2    nstate             // two temporary
                                 //nstates。z1指向当前状态，z2临时存储后继状态
}

// A dstate corresponds to a DFA state.
type dstate struct {
    next     [256]*dstate // next state, per byte
    enc      string       // encoded nstate
    matchNL  bool         // match when next byte is \n
    matchEOT bool         // match in this state at end of text
}

// An nstate corresponds to an NFA state.
type nstate struct {
    q       sparse.Set // queue of program instructions
    partial rune       // partially decoded rune (TODO)
    flag    flags      // flags (TODO)
}

// A Set is a sparse set of uint32 values.
// http://research.swtch.com/2008/03/using-uninitialized-memory-for-fun-and.html
type sparse.Set struct {
    dense  []uint32    // 真正存储的数组
    sparse []uint32    // 对应数字在dense中的位置
}
```


## index.Query
```go
// A Query is a matching machine, like a regular expression,
// that matches some text and not other text.  When we compute a
// Query from a regexp, the Query is a conservative version of the
// regexp: it matches everything the regexp would match, and probably
// quite a bit more.  We can then filter target files by whether they match
// the Query (using a trigram index) before running the comparatively
// more expensive regexp machinery.
type Query struct {
    Op      QueryOp
    Trigram []string
    Sub     []*Query
}
```


## regexp.Grep
```go
// TODO:
type Grep struct {
    Regexp *Regexp   // regexp to search for
    Stdout io.Writer // output target
    Stderr io.Writer // error target

    L bool // L flag - print file names only
    C bool // C flag - print count of matches
    N bool // N flag - print line numbers
    H bool // H flag - do not print file names

    Match bool

    buf []byte
}
```