# 正则表达式与RE2


## 确定有限状态机与非确定有限状态机
非确定有限状态机，遇到同一输入，有不同的后继状态，而有限状态机，同一的输入，只有一个后继状态

## 正则表达式与NFA
每一个正则表达式都可以转换为唯一的NFA，并且NFA中的状态树最多等于原始正则表达式的长度

## 正则表达式搜索算法

### 非确定有限状态机处理
当存在状态S，遇到同一输入有多个后继状态时，处理方式有两种：
1. 逐个进入每个后继状态，当不匹配时，状态回退到S，尝试下一个状态
2. 同时进入每个状态，当不匹配时丢弃不匹配的状态分支，匹配的继续
第一种方法的缺点在于需要多次回退，在成功匹配前可能会多次读取字符串。而第二种方法仅不需要进行回退，可以保证状态机的行为在线性时间下完成。

## nfa的生成与执行
[nfa文档](https://swtch.com/~rsc/regexp/regexp1.html)
[nfa源码](https://swtch.com/~rsc/regexp/nfa.c.txt)

正则表达式经过转换为nfa图，为需要匹配的字符串，提供移动方向
```c
char*
re2post(char *re) // 将正则表达式按照一定的规则转换为后缀表达式

State*
post2nfa(char *postfix) // 将后缀表达式转换为可用于执行的nfa，nfa为由结构体State组成的树。nfa树的叶子节点为特殊状态，表示成功匹配
```

match函数互将遍历一遍字符串，并按字符寻找start代表的nfa中的后继状态。step函数会在clist中寻找将满足字符c的所有后继状态，并且将其保存在nlist。如果在step中没有查找到后继状态，那么对应的nlist为空。match中读取下一字符后step会使得nlist永远保持为空，将导致ismatch返回假值。
```c
struct State
{
    int c;
    State *out;
    State *out1;
    int lastlist;
};

int
match(State *start, char *s)
{
    List *clist, *nlist, *t;

    /* l1 and l2 are preallocated globals */
    clist = startlist(start, &l1);
    nlist = &l2;
    for(; *s; s++){
        step(clist, *s, nlist);
        t = clist; clist = nlist; nlist = t;    /* swap clist, nlist */
    }
    return ismatch(clist);
}

void
step(List *clist, int c, List *nlist)
{
    int i;
    State *s;

    listid++;
    nlist->n = 0;
    for(i=0; i<clist->n; i++){
        s = clist->s[i];
        if(s->c == c)
            addstate(nlist, s->out);
    }
}

int
ismatch(List *l)
{
    int i;

    for(i=0; i<l->n; i++)
        if(l->s[i] == matchstate)
            return 1;
    return 0;
}
```

## nfa执行转dfa执行
[dfa文档](https://swtch.com/~rsc/regexp/regexp1.html)
[dfa源码](https://swtch.com/~rsc/regexp/dfa0.c.txt)

在资料中通过DState结构保存了每个状态遇到输入后的直接后继,减少对DState.l的重复遍历计算。

dfa的执行是通过nextstate函数完成对nfa的步进与dfa的构建，dfa的构建是在字符串进行正则匹配，并且遇到相关字符才会进行对应的后继DState构建
```c
struct DState
{
	List l;					// 保存对应状态的后继状态列表
	DState *next[256];     // 保存对应字符的后继状态或DState
	DState *left;          // 状态的左侧，用于指向新增DState应在整个结构中保存的结构
	DState *right;         // 状态的右侧，用于保存新增DState在整个结构中的位置
};

int
match(DState *start, char *s)
{
    int c;
    DState *d, *next;
    
    d = start;
    for(; *s; s++){
        c = *s & 0xFF;
        if((next = d->next[c]) == NULL)
            next = nextstate(d, c);
        d = next;
    }
    return ismatch(&d->l);
}

DState*
dstate(List *l)
{
    int i;
    DState **dp, *d;
    static DState *alldstates;

    qsort(l->s, l->n, sizeof l->s[0], ptrcmp);

    /* look in tree for existing DState */
    dp = &alldstates;
    while((d = *dp) != NULL){
        i = listcmp(l, &d->l);
        if(i < 0)
            dp = &d->left;
        else if(i > 0)
            dp = &d->right;
        else
            return d;
    }
    
    /* allocate, initialize new DState */
    d = malloc(sizeof *d + l->n*sizeof l->s[0]);
    memset(d, 0, sizeof *d);
    d->l.s = (State**)(d+1);
    memmove(d->l.s, l->s, l->n*sizeof l->s[0]);
    d->l.n = l->n;

    /* insert in tree */
    *dp = d;
    return d;
}
```

当DState.next[c]不存在时，表示该状态下的字符c尚未形成对应的状态，此时会遍历其后继状态列表，并生成next[c]的值
```
DState*
nextstate(DState *d, int c)
{
    step(&d->l, c, &l1);
    return d->next[c] = dstate(&l1);
}
```