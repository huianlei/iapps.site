# iapps.site
A practise project written by golang
## 概述

	本程序服务器和客户端使用Golang开发。需求见下文【开发要求】部分
	程序流程说明：
	服务器监听Tcp端口（默认9001），接受客户端连接。
	客户端连接服务器后，发起CGLoginMessage登录消息，服务器收到登录消息后，将消息push 到QueueService的消息队列中(channel)，实现登录排队
	同时，负责账号Token验证的多个goroutine，从channel中读取登录消息，验证（最小化模拟sleep 100 ms）成功后，添加到PlayerManager中，表示玩家登录成功。
	
	服务器保护：
	队列channel有容量限制，超出后，再请求登录的客户端将直接提示 queue full。 服务器主动断开连接。
	最大同时在线人数限制，超出后，不在消费等待队列中的数据，除非有在线玩家退出（时间关系暂未模拟在线玩家退出）
