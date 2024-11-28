### 在用masscan工具进行端口扫描的时候遇到的问题
1. 在程序执行的时候，如果加了参数-oJ -，stdout管道输出的就是扫描到的记录的json格式，详情见stdout.log
2. stderr管道输出的是扫描进度信息，详情见stderr.log
3. 对于stdout，我是启动了一个协程来处理标准输出的，其中我新建了一个scanner来读取stdout管道的内容(`scanner := bufio.NewScanner(stdout)`)，对于符合条件的记录解析成json，然后提取出其中的(传输层协议, IP, Port) 没什么问题
4. 但是对于stderr，我也启动了一个协程来处理错误输出，也是新建了一个scanner来读取stderr管道的内容(`scanner := bufio.NewScanner(stderr)`)，对于每条进度日志，用正则提取出其中的进度百分比，存入Redis，但是出现了一个问题：程序运行时用`redis-cli monitor`查看Redis语句记录没有任何动静，只有在程序结束或者我手动停止程序的时候才会插入记录语句
5. 排查到是scanner的问题，换成了buffer（暂时还不知道为啥）