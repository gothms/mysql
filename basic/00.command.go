package basic

/*
cmd
	mysql -u root -p / mysql -h localhost -P 3306 -u root -p
		用户 root 登录
	输入密码
	show databases;
	use 表名;
	show tables
	quit / exit


mysqlsh
	mysql -h localhost -P 3306 -u BeenLee -p
	mysql -h 127.0.0.1 -P 3306 -u BeenLee -p
		报错
			SyntaxError: Unexpected identifier
		解决
			\connect root@localhost
		用户名/密码错误
			Access denied for user
*/
