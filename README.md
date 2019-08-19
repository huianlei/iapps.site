# iapps.site
A practise project written by golang

## 概述

* 本程序服务器和客户端使用Golang开发。需求见工程readme.txt中【开发要求】部分
* 程序流程说明：
服务器监听Tcp端口（默认9001），接受客户端连接。
客户端连接服务器后，发起CGLoginMessage登录消息，服务器收到登录消息后，将消息push 到QueueService的消息队列中(channel)，实现登录排队同时，负责账号Token验证的多个goroutine，从channel中读取登录消息，验证（最小化模拟sleep 100 ms）成功后，添加到PlayerManager中，表示玩家登录成功。
	
* 服务器保护：
队列channel有容量限制，超出后，再请求登录的客户端将直接提示 queue full。 服务器主动断开连接。最大同时在线人数限制，超出后，不在消费等待队列中的数据，除非有在线玩家退出（时间关系暂未模拟在线玩家退出）

## 环境要求
Windows/Linux <br /> 
Go1.9+ <br />

## 安装部署
部署前请正确安装Golang，并正确设置Path环境变量，确保go version 命令运行OK。<br />
如有疑问请 google golang 安装，并请正确设置GOPATH环境变量。<br />
以笔者GOPATH=E:\develop\golang 为例<br />
<code>
	
	mkdir -p E:\develop\golang\src
	cd /d E:\develop\golang\src

</code>
将本工程下载到src目录下，则 E:\develop\golang\src\iapps.site，此目录即为程序脚本工作目录

## 如何使用
* on Windows
** 运行服务器
双击 win_server_start.bat 运行服务器<br />
或者命令行模式下，cd 当前目录输入以下命令：<br />
go run app.go 
** 运行客户端
双击 win_client_start.bat 运行压测客户端（如需调整参数，请自行修改，或者直接运行下面的命令）<br />
或者命令行模式下，cd 当前目录输入以下命令:（开启10000个客户端连接，每秒并发200个）<br />
<code>
	
	go run app.go -connect localhost:9001 -count 10000 -concurrent 200
	
</code>
* on Linux
<code>
	
	cd iapps.site <br />
	sh server_start.sh  // 启动服务器前台运行，Ctrl+C 终止运行 <br />

</code>
注意：上述脚本为先编译源码，再运行方式。方便修改配置参数后，重新运行。

## 配置常量说明
常量配置文件目录： iapps.site/common/common.go <br/>
可根据压测数据调整下列参数。
<code>
	
	// PlayerManager const config
	const (
		// max online player count 
		// 最大同时在线人数,达到最大在线后，不再消费登录队列消息，除非有玩家退出（这个时间关系，尚未模拟）
		MaxOnline int32 = 1000
	)
	
	// QueueService const config
	const (
		// tick milliseconds
		TickInterval = int64(100)
		// queue service channel capacity  
		// 登录排队队列容量，超出容量则直接给客户端提示 queue full
		QueueCapacity = int(500)
		// check interval to broadcast position in queue. in seconds
		// 登录队列扫描时间间隔：秒。对于排队的客户端，发生变化后，定期同步给客户端排队的位置
		QueueCheckInterval = int64(3)
	)

	// TcpServer const
	const (
		// Listen port
		Port = ":9001"
		// simulate validate token sleep time in milliseconds
		// 模拟登录Token校验的时间消耗：毫秒
		ValidateTokenSleep = int64(100)
	)
	
</code>

## 参考资料
<p><a href="https://tour.go-zh.org/list">Go 指南</a></p>
<p><a href="https://segmentfault.com/a/1190000014733620">Go语言TCP/IP网络编程</a></p>
<p><a href="https://tiancaiamao.gitbooks.io/go-internals/content/zh/">深入解析Go</a></p>
