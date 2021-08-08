package main

import (
	"mysql-proxy/lib"
	"net"
	"runtime"
	"sync"
)

func main() {

	mysqlUrl := ":3306" // 数据库的端口
	listen := ":8899"   // 对外暴露的端口
	connCount := runtime.NumCPU()
	//connCount := 1 // local 只是开启一个 , navicat 是开启多个连接的,开启一个有问题

	server, _ := net.Listen("tcp", listen)
	connList := make([]lib.ProxyConn, connCount)
	for i := 0; i < connCount; i++ {
		proxy := lib.ProxyConn{}
		connList[i] = proxy
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	for i := 0; i < connCount; i++ {

		go func(proxy lib.ProxyConn) {
			// 一: client , mysql的连接信息
			// 初始化mysql client
			proxy.NewClientConn(server)
			// 初始化mysql server
			proxy.NewMysqlConn(mysqlUrl)
			// 类似于登录过程,换token存储于:FinishHandshakePacket  , auth操作
			err := proxy.Handshake()
			if err != nil {
				proxy.Close()
			}
			// 读取mysql server->client
			go proxy.PipeMysql2Client()
			// 读取mysql  client->server
			go proxy.PipeClient2Mysql()

			for {
				// 二: 请求客户端的连接
				// todo 待优化
				//if !proxy.IsClientClose() {
				//	fmt.Println("xxx", time.Now())
				//	continue
				//}
				proxy.NewClientConn(server)
				// TODO 这个也需要再次auth吗
				//err := proxy.FakeHandshake()
				//if err != nil {
				//	proxy.CloseClient()
				//}
				go proxy.PipeClient2Mysql()
			}

		}(connList[i])

	}

	wg.Wait()
}
