# cercis-server

## 部署说明

### native 部署

后端使用 golang + redis + postgresql 的技术栈。

`cercis-server/` 的目录组织结构如下：

```
cercis-server
├── api        http(s) 的 api 交互层
│   ├── activity    动态相关 api
│   ├── auth        用户验证相关 api
│   ├── chat        聊天相关 api
│   ├── friend      好友相关 api
│   ├── mobile      手机短信相关 api
│   ├── search      搜索相关 api
│   ├── upload      图片/视频上传相关 api
│   └── user        用户相关 api
├── config     配置文件读取
├── db         数据库交互层
│   ├── activity    动态相关 dao
│   ├── chat        聊天相关 dao
│   └── user        用户相关 dao
├── logger     日志模块
├── middleware 中间件
├── redis      redis 交互模块
├── util       某些工具
│   ├── security    安全相关工具
│   ├── sms         短信服务工具
│   └── validator   参数合法性校验工具
└── ws         websocket 相关
```

项目根目下的 `config-template.yml` 文件中，指定了后端运行过程中所需要的各类配置，包括但不限于数据库连接、sms 服务、七牛云图床服务和服务器运行参数等。

在配置文件完善后，使用 `go run -c <config-path>` 运行项目即可。

### docker-compose 部署

在我们实际的服务器上，已经通过 docker-compose 工具部署了后端，可以根据自己的需求调整 `docker-compose.yml` 中的内容。其中信息大致为：

```yaml
version: "3"
services:
  cercis_postgres:
    image: postgres:13.2-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=pg_cercis
      - POSTGRES_DB=cercis
    volumes:
      - "./data:/var/lib/postgresql/data"
    restart: always
  cercis_redis:
    image: redis:6.2.2-alpine
    volumes:
      - "./redis:/data"
    restart: always
  cercis:
    image: golang:1.16.3-alpine
    working_dir: "/src"
    environment:
      - GOPROXY=https://mirrors.aliyun.com/goproxy/
    ports:
      - "9191:9191"
    depends_on:
      - cercis_postgres
      - cercis_redis
    restart: always
    command: ["go", "run", "main.go", "-c", "./config.yml"]
    volumes:
      - "./cercis-server:/src"
```

在此 `docker-compose.yml` 中，运行了 redis + postgresql　数据库，将 `./cercis-server`　文件夹挂载到容器中，并运行配置文件 `config.yml`。

### config.yml 详解

`config-template.yml` 中列出了后端运行所需要的配置：

```yaml
server:
  host: "0.0.0.0"    # 后端
  port: 9191         # 后端监听端口
  logger:
    # 0-Panic, 1-Fatal, 2-Error, 3-Warn, 4-Info, 5-Debug, 6-Trace
    level: 4  # logger 输出等级。
redis:
  host: "127.0.0.1"  # redis 运行地址
  port: 6379         # redis 监听端口
  username: ""       # redis 用户名
  password: ""       # redis 密码
  # 选择数据库
  database: 7        # redis 使用的数据库
  # 连接时是否重置
  reset: false       # 连接时是否清空原有数据
postgres:
  host: "localhost"     # pg 运行地址
  port: 5432            # pg 监听端口
  user: "postgres"      # pg 用户名
  password: "123456"  # pg 密码
  dbname: "cercis"      # 数据库名称
  sslmode: "disable"
  timezone: "Asia/Shanghai" # 时区
# SMS 短信服务，目前使用阿里云
sms:
  region: "cn-beijing"    # 阿里云 sms 区域
  # 下面为阿里云 sms 服务专属配置，详情请见相应资料
  accesskey: "<aliyun-sms-accesskey>"
  secret: "<aliyun-sms-secret>"
  signname: "幻想乡"
  templatecode: "<sms-template-code>"
# 七牛云对象存储服务，详情请见相应资料
qiniu:
  accesskey: ""
  secretkey: ""
  bucket: "cercis"

```