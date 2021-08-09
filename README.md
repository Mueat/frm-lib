### golib

golang常用库

### 库

- [x] config 配置文件读取库
- [x] errors 错误码设置
- [x] log 日志
- [x] db 数据库
- [x] cache 缓存
- [x] crypt  加密库
- [x] utils  常用工具

### config

配置文件为toml格式的文件

```go
import "github.com/Mueat/golib/config"

type Config struct {
	AppName string
	Basic   BasicConfig
	Log     map[string]LogConfig
	Mysql   map[string]MysqlConfig
	Redis   map[string]RedisConfig
}

func main() {
	conf := Config{}
	err := config.ParseConfig("./config.toml", &conf)
	if err != nil {
		panic(err)
	}
}
```

### errors

错误定义文件格式如下

```go
package errors

const (
	// success
	OK = 0
	// 系统错误
	System = 1
	// 参数错误
	Params = 2
)
```

使用解析方法生成错误字典文件

```go
errors.ParseErrors("./errors/errors.go", "./errors/errors_map.go", "errors")
```


### TODO

- [ ] 新增链路追踪
- [ ] db类中自定义logger实现链路追踪
- [ ] redis类实现链路追踪
- [ ] curl 实现链路追踪
- [ ] app 新增链路追踪context