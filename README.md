# eslogrus —— an es hook for logrus

一个简单的基于[elasticsearch](github.com/elastic/go-elasticsearch)的钩子，对[logrus](github.com/sirupsen/logrus)日志包进行收集作用

# 主要功能

> 慢慢完善吧...

- [x] 日志来一条发一条
- [ ] 累计日志条数，再发送es
- [ ] 累计时间定时发送es
- [ ] 支持tls

# 注意
启动es，可以按照一下给出的docker-compose来启动,因为最新版本的es需要tls来支持数据安全，而我们这里没有加，所以就要手动设置`xpack.security.enabled`

```yaml
version: '3.7'

services:
  elasticsearch:
    image: elasticsearch:8.4.2
    container_name: elasticsearch
    environment:
      bootstrap.memory_lock: true
      ES_JAVA_OPTS: "-Xms512m -Xmx512m"
      discovery.type: single-node
      ingest.geoip.downloader.enabled: false
      TZ: Asia/Shanghai
      xpack.security.enabled: false
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:9200/_cluster/health || exit 1"] #⼼跳检测，成功之后不再执⾏后⾯的退出
      interval: 60s #⼼跳检测间隔周期
      timeout: 10s
      retries: 3
      start_period: 60s #⾸次检测延迟时间
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - /usr/local/elasticsearch/data:/usr/local/elasticsearch/data
      - /usr/local/elasticsearch/config/es/config:/usr/local/elasticsearch/config
    ports:
      - "9200:9200"
    restart: always
```

# 用法

- 初始化ES

```go
var esClient *elastic.Client

func InitEs() {
	cfg := elastic.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	client, err := elastic.NewClient(cfg)
	if err != nil {
		log.Panic(err)
	}
	esClient = client
}
```

- 初始化日志对象

```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
logger.SetFormatter(&logrus.JSONFormatter{
    TimestampFormat: "2006-01-02 15:04:05",
})
// 其他设置...
```

- 新建一个hook

```go
hook, err := NewElasticHook(esClient, "localhost", logrus.DebugLevel, "my_index")
if err != nil {
    fmt.Println("err", err)
}
```

- 日志对象添加hook

```go
logger.AddHook(hook)
logger.Error("这是一个测试情况")
```

- 查询结果

```shell
curl --location 'http://localhost:9200/my_index/_search'\?pretty
```

详细可以看test文件