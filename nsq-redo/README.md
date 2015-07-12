# nsq-redo
collect redo logs for player data change information
#Redo Logs原理
在数据更新操作commit前，将更改的SQL脚本写入重做日志。主要用于数据库的增量备份和增量恢复。

