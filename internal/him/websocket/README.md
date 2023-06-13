## 服务端server.go
1. 握手升级，HTTP --> websocket
2. newConn(net.conn) 将net.Conn包装成him.Conn
3. r.Accept(conn, r.options.loginWait) 回调到上层业务完成权限认证
4. r.Add(channel) 自动添加到him.ChannelMap连接管理器
5. ch.ReadLoop(r.MessageListener) 开启一个goroutine中循环读取消息

## 客户端client.go
1. Connect：拨号建立并连接握手
2. heartbeatLoop：发送心跳
3. Read：读取消息
4. Send：发送消息