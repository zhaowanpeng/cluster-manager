提供以下功能：

1. 添加node，user，pwd

参数：
-g
-n
-o 默认：22，端口
-u 默认：root，用户
-b 默认：true,合并输出
-t 默认：20，等待时间

clush add -g group1 -n 192.168.108.1-100 -port 22 -u root -d "这是xx" -p
input password:***隐藏输入***


table: shell_client
id: string
ip: string
port: int
user: string
password: string
group: string
add_at: datetime
usable: bool
tmp: bool 是否是响应不同划分出来的临时分组
description: string 默认""

如果group,ip,port,user已经存在，更新password和add_at
go协程ssh验证每个用户是否可以登录，可以登陆的usable = true

当默认-b时分组返回
----------------------------------------
tmp-随机值-1(1) 192.168.1.6：
登录成功!
tmp-随机值-2(89) 192.168.1.1-5,192.168.1.7-90:
登录失败!
tmp-随机值-3(10) 192.168.1.91-100:
connect timeout!

根据不同的响应结果，随机生成n个临时分组，
分组名为tmp-随机值，description为：
output: xxx
保存到数据库


2. 删除分组 或 分组中的某个特定值

参数
-g 必填
-u * 默认全部
-o * 默认全部
-n * 默认全部

clush delete -g group1 -n 192.168.108.1-3 -u root -o 22

clush delete -g group1


3. 查看分组列表

clush list
group-name1(192.168.108.1-100)
group-name2(192.168.108.1-100,192.168.108.200)


4. 在特定分组执行命令

单次：
clush exec -g group1 "command"
clush exec -n 192.168.1.2-5,192.168.1.9 "command"

多次执行：
clush exec -g group1
clush exec -n 192.168.1.2-5,192.168.1.9
>>>"uname -a"
g1() output:
aaa
g2() output:
bbb
>>>"uname -a"
g1() output:
aaa
g2() output:
bbb

临时：
clush exec -n 192.168.1.2-5,192.168.1.9



4.分发文件/文件夹
clush scp -g group1 /root/a/haha.txt /root/b/11.txt
clush scp -g group1 /root/a/haha.txt /root/b/
clush scp -g group1 -r /root/a /root/b/


5.收集文件/文件夹
clush recv -g group1 -p /root/a/app.log -











2. 设置分组



1.接口，接收个节点pub或者密码

1，集群账户、密码、权限管理
2，分发指令
3，分发文件
4，分发会话
5，安装步骤保存，分发到不同机器