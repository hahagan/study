https://github.com/torvalds/linux/blob/master/Documentation/filesystems/overlayfs.rst

https://arkingc.github.io/2017/08/18/2017-08-18-linux-code-vfs/#

### 合并策略

将底层数据层合并为只读的lower层，并新建可写的upper层，对外提供的两层联合视图为merge层，对merge层的操作会转变为对upper和lower的操作。



### 打开目录

```shell
$strace ls merged/
......
......
stat("merged/", {st_mode=S_IFDIR|0755, st_size=6, ...}) = 0
openat(AT_FDCWD, "merged/", O_RDONLY|O_NONBLOCK|O_CLOEXEC|O_DIRECTORY) = 3	## 获取文件对象
getdents(3, /* 23 entries */, 32768)    = 600
getdents(3, /* 0 entries */, 32768)     = 0
close(3)                                = 0
fstat(1, {st_mode=S_IFCHR|0620, st_rdev=makedev(136, 0), ...}) = 0
mmap(NULL, 4096, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7f8246087000
write(1, "anaconda-post.log  bin\tdev  etc "..., 124anaconda-post.log  bin	dev  etc  home	lib  lib64  lost+found	media  mnt  opt  proc  root  run  sbin	srv  sys  tmp  usr  var
) = 124
close(1) 
```



### 数据结构

文件描述符

```c
struct file {
	union {
		struct llist_node	fu_llist;
		struct rcu_head 	fu_rcuhead;
	} f_u;
	struct path		f_path;
	struct inode		*f_inode;	/* cached value */
	const struct file_operations	*f_op;

	/*
	 * Protects f_ep, f_flags.
	 * Must not be taken from IRQ context.
	 */
	spinlock_t		f_lock;
	enum rw_hint		f_write_hint;
	atomic_long_t		f_count;
	unsigned int 		f_flags;
	fmode_t			f_mode;
	struct mutex		f_pos_lock;
	loff_t			f_pos;
	struct fown_struct	f_owner;
	const struct cred	*f_cred;
	struct file_ra_state	f_ra;

	u64			f_version;
#ifdef CONFIG_SECURITY
	void			*f_security;
#endif
	/* needed for tty driver, and maybe others */
	void			*private_data;

#ifdef CONFIG_EPOLL
	/* Used by fs/eventpoll.c to link all the hooks to this file */
	struct hlist_head	*f_ep;
#endif /* #ifdef CONFIG_EPOLL */
	struct address_space	*f_mapping;
	errseq_t		f_wb_err;
	errseq_t		f_sb_err; /* for syncfs */
} __randomize_layout
  __attribute__((aligned(4)));	/* lest something weird decides that 2 is OK */
```

