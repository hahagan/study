go标准正则库.md

# 重要数据结构
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

// A Prog is a compiled regular expression program.
type Prog struct {
    Inst   []Inst
    Start  int // index of start instruction
    NumCap int // number of InstCapture insts in re
}

// An Inst is a single instruction in a regular expression program.
type Inst struct {
    Op   InstOp
    Out  uint32 // all but InstMatch, InstFail
    Arg  uint32 // InstAlt, InstAltMatch, InstCapture, InstEmptyWidth
    Rune []rune
}

const (
    InstAlt InstOp = iota
    InstAltMatch
    InstCapture
    InstEmptyWidth
    InstMatch
    InstFail
    InstNop
    InstRune
    InstRune1
    InstRuneAny
    InstRuneAnyNotNL
)


const (
    OpNoMatch        Op = 1 + iota // matches no strings
    OpEmptyMatch                   // matches empty string
    OpLiteral                      // matches Runes sequence
    OpCharClass                    // matches Runes interpreted as range pair list
    OpAnyCharNotNL                 // matches any character except newline
    OpAnyChar                      // matches any character
    OpBeginLine                    // matches empty string at beginning of line
    OpEndLine                      // matches empty string at end of line
    OpBeginText                    // matches empty string at beginning of text
    OpEndText                      // matches empty string at end of text
    OpWordBoundary                 // matches word boundary `\b`
    OpNoWordBoundary               // matches word non-boundary `\B`
    OpCapture                      // capturing subexpression with index Cap, optional name Name
    OpStar                         // matches Sub[0] zero or more times
    OpPlus                         // matches Sub[0] one or more times
    OpQuest                        // matches Sub[0] zero or one times
    OpRepeat                       // matches Sub[0] at least Min times, at most Max (Max == -1 is no limit)
    OpConcat                       // matches concatenation of Subs
    OpAlternate                    // matches alternation of Subs
)

type parser struct {
    flags       Flags     // parse mode flags
    stack       []*Regexp // stack of parsed expressions
    free        *Regexp
    numCap      int // number of capturing groups seen
    wholeRegexp string
    tmpClass    []rune // temporary char class work space
}

const (
    FoldCase      Flags = 1 << iota // case-insensitive match
    Literal                         // treat pattern as literal string
    ClassNL                         // allow character classes like [^a-z] and [[:space:]] to match newline
    DotNL                           // allow . to match newline
    OneLine                         // treat ^ and $ as only matching at beginning and end of text
    NonGreedy                       // make repetition operators default to non-greedy
    PerlX                           // allow Perl extensions
    UnicodeGroups                   // allow \p{Han}, \P{Han} for Unicode group and negation
    WasDollar                       // regexp OpEndText was $, not \z
    Simple                          // regexp contains no counted repetition

    MatchNL = ClassNL | DotNL

    Perl        = ClassNL | OneLine | PerlX | UnicodeGroups // as close to Perl as possible
    POSIX Flags = 0                                         // POSIX syntax
)
```

解析正则表达式主要函数
```go
// literal pushes a literal regexp for the rune r on the stack
// and returns that regexp.
func (p *parser) literal(r rune) {
    p.push(p.newLiteral(r, p.flags))
}

// op pushes a regexp with the given op onto the stack
// and returns that regexp.
func (p *parser) op(op Op) *Regexp {
    re := p.newRegexp(op)
    re.Flags = p.flags
    return p.push(re)
}
```

# 正则表达式的解析
主要函数
```go
// alternate replaces the top of the stack (above the topmost '(') with its alternation.
func (p *parser) alternate() *Regexp {
    // Scan down to find pseudo-operator (.
    // There are no | above (.
    i := len(p.stack)
    for i > 0 && p.stack[i-1].Op < opPseudo {
        i--
    }
    subs := p.stack[i:]
    p.stack = p.stack[:i]

    // Make sure top class is clean.
    // All the others already are (see swapVerticalBar).
    if len(subs) > 0 {
        cleanAlt(subs[len(subs)-1])
    }

    // Empty alternate is special case
    // (shouldn't happen but easy to handle).
    if len(subs) == 0 {
        return p.push(p.newRegexp(OpNoMatch))
    }

    return p.push(p.collapse(subs, OpAlternate))
}

// concat replaces the top of the stack (above the topmost '|' or '(') with its concatenation.
func (p *parser) concat() *Regexp {
    p.maybeConcat(-1, 0)

    // Scan down to find pseudo-operator | or (.
    i := len(p.stack)
    for i > 0 && p.stack[i-1].Op < opPseudo {
        i--
    }
    subs := p.stack[i:]
    p.stack = p.stack[:i]

    // Empty concatenation is special case.
    if len(subs) == 0 {
        return p.push(p.newRegexp(OpEmptyMatch))
    }

    return p.push(p.collapse(subs, OpConcat))
}
```

以下分析中以正则表达式"(test+|tee+)"为输入
0. 正则表达式经过parse函数解析，最终会形成一棵由正则表达式生成的语法树

1. regexp.parse逐个读入字符，大多数输入，会直接形成特定的Regexp节点，并存入stack数组中。
    遇到字符输入时，会项数组内添加形成的rexgexp之前，会对数组内已存在的多个literal类型的rexgexp进行合并，将可合并项合并为一个regexp。
    例如当前p.stack为["(","t","e"]，那么当分析到字符"s"时会将数据前两个字符进行合并最终变为["(","te","s"]

2. 当遇到重复符"+","?","*"或者一些特殊输入时，会将p.stack中的最后一个(或多个)regexp存入新生成的regexp的子regexp数组中(可以认为是新regexp的子节点)。
```
    例如当前p.stack为["(","tes","t"],此时会遇到符号"+",那么会将"t"存入新生成的plus类型的regexp的sub对象中。即p.stack为
        ["(", "tes", plus, alternate]
                       |
                     ["t"] 
```

3. 当遇到"|"符号时，会处理当前p.stack中可以"concat"的Regexup，并处理已包含的"|"的Regexp，如果尚未存在"|"的regexp，形成"|"的regexp并存入p.stack
```
    例如当输入为"|"时，p.stack会被处理为
    ["(", concat, alternate]
           /         
    ["tes", plus]
              |
            ["t"]
```

4. 当输入结束后或遇到某些特殊字符时，会对p.stack中的regexp进行最终的"concat"和"alternate"。例如右括号")"在处理前，需要对当前p.stack中可合并项的合并和分支处理
```
    当输入为")"时，未处理前p.stack为
    ["(", concat, alternate, concat]
           /                    \
    ["tes", plus]            ["tee", plus]
              |                        |
            ["t"]                    ["t"]

    将可连接项合并后，随后会通过"alternate"生成分支，同时会进行一些同项合并，变为

    ["(", concat]
           /
        ["te", alternate]
                /
             [concat,  concat]
              /             \
        ["s", plus]        ["e", plus]
                |                  |
              ["t"]              ["t"]

    最后再进行")"相关的处理,p.stack变为：
       [capture]
           /
        ["te", alternate]
                /
             [concat,  concat]
              /             \
        ["s", plus]        ["e", plus]
                |                  |
              ["t"]              ["t"]

   如果本例中输入"(test+|test+)"去掉左右括号变为"test+|teet+"，那么最终的"concat","alternate"在输入遍历结束后进行
        [concat, alternate, concat]
           /                    \
    ["tes", plus]            ["tee", plus]
              |                        |
            ["t"]                    ["t"]
    遍历结束后变为
    ["te", alternate]
                /
             [concat,  concat]
              /             \
        ["s", plus]        ["e", plus]
                |                  |
              ["t"]              ["t"]
```