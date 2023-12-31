package basic

/*
日志系统：一条SQL更新语句是如何执行的？

MySQL 可以恢复到半个月内任意一秒的状态，这是怎样做到的呢？

示例
	创建表T，有一个主键 ID 和一个整形字段 c
		mysql> create table T(ID int primary key, c int);
	将 ID=2 这一行的值加 1
		mysql> update T set c=c+1 where ID=2;
流程
	连接数据库，这是连接器的工作
	在一个表上有更新的时候，跟这个表有关的查询缓存会失效，所以这条语句就会把表 T 上所有缓存结果都清空
		这也就是我们一般不建议使用查询缓存的原因
	分析器会通过词法和语法解析知道这是一条更新语句
	优化器决定要使用 ID 这个索引
	执行器负责具体执行，找到这一行，然后更新
	与查询流程不一样的是，更新流程还涉及两个重要的日志模块
		redo log（重做日志）和 binlog（归档日志）

重要的日志模块：redo log
	举例：《孔乙己》
		酒店掌柜有一个粉板，专门用来记录客人的赊账记录
		如果有人要赊账或者还账的话，掌柜一般有两种做法：
			一种做法是直接把账本翻出来，把这次赊的账加上去或者扣除掉
			另一种做法是先在粉板上记下这次的账，等打烊以后再把账本翻出来核算
	MySQL 需求分析
		如果每一次的更新操作都需要写进磁盘，然后磁盘也要找到对应的那条记录，然后再更新，整个过程 IO 成本、查找成本都很高
	WAL 技术（Write-Ahead Logging）：关键点就是先写日志，再写磁盘
		当有一条记录需要更新的时候，InnoDB 引擎就会先把记录写到 redo log（粉板）里面，并更新内存，这个时候更新就算完成了
		InnoDB 引擎会在适当的时候，将这个操作记录更新到磁盘里面，而这个更新往往是在系统比较空闲的时候做
	redo log
		InnoDB 的 redo log 是固定大小的，比如可以配置为一组 4 个文件，每个文件的大小是 1GB，那么这块“粉板”总共就可以记录 4GB 的操作
		从头开始写，写到末尾就又回到开头循环写
		如图 02.update_redo_log.jpg
			write pos 是当前记录的位置，一边写一边后移，写到第 3 号文件末尾后就回到 0 号文件开头
			checkpoint 是当前要擦除的位置，也是往后推移并且循环的，擦除记录前要把记录更新到数据文件
			write pos 和 checkpoint 之间的是“粉板”上还空着的部分，可以用来记录新的操作
			如果 write pos 追上 checkpoint，表示“粉板”满了，这时候不能再执行新的更新，得停下来先擦掉一些记录，把 checkpoint 推进一下
	crash-safe
		有了 redo log，InnoDB 就可以保证即使数据库发生异常重启，之前提交的记录都不会丢失，这个能力称为crash-safe
重要的日志模块：binlog
	简介
		MySQL 整体来看，其实就有两块：一块是 Server 层，它主要做的是MySQL 功能层面的事情；还有一块是引擎层，负责存储相关的具体事宜
		redo log 是 InnoDB 引擎特有的日志
		而 Server 层也有自己的日志，称为binlog（归档日志）
	为什么会有两份日志呢？
		因为最开始 MySQL 里并没有 InnoDB 引擎。MySQL 自带的引擎是 MyISAM，但是 MyISAM 没有 crash-safe 的能力，binlog 日志只能用于归档
		而 InnoDB 是另一个公司以插件形式引入 MySQL 的，既然只依靠 binlog 是没有 crash-safe 能力的
			所以 InnoDB 使用另外一套日志系统——也就是 redo log 来实现 crash-safe 能力
	redo log VS binlog
		1.redo log 是 InnoDB 引擎特有的
			binlog 是 MySQL 的 Server 层实现的，所有引擎都可以使用
		2.redo log 是物理日志，记录的是“在某个数据页上做了什么修改”
			binlog 是逻辑日志，记录的是这个语句的原始逻辑，比如“给 ID=2 这一行的 c 字段加 1 ”。
		3.redo log 是循环写的，空间固定会用完
			binlog 是可以追加写入的。“追加写”是指 binlog 文件写到一定大小后会切换到下一个，并不会覆盖以前的日志

update 内部流程
	1. 执行器先找引擎取 ID=2 这一行。ID 是主键，引擎直接用树搜索找到这一行
		如果 ID=2 这一行所在的数据页本来就在内存中，就直接返回给执行器
		否则，需要先从磁盘读入内存，然后再返回
	2. 执行器拿到引擎给的行数据，把这个值加上 1，比如原来是 N，现在就是 N+1，得到新的一行数据，再调用引擎接口写入这行新数据
	3. 引擎将这行新数据更新到内存中，同时将这个更新操作记录到 redo log 里面，此时 redo log 处于 prepare 状态
		然后告知执行器执行完成了，随时可以提交事务
	4. 执行器生成这个操作的 binlog，并把 binlog 写入磁盘
	5. 执行器调用引擎的提交事务接口，引擎把刚刚写入的 redo log 改成提交（commit）状态，更新完

图中浅色框表示是在 InnoDB 内部执行的，深色框表示是在执行器中执行的







*/
