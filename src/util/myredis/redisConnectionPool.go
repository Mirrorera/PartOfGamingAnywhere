package myredis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/astaxie/beego"
	"time"
)

var (
	// 定义常量
	RedisClient     *redis.Pool
	REDIS_HOST string
	REDIS_DB   int
)

func Init() {
	// 从配置文件获取redis的ip以及db
	beego.LoadAppConfig("ini", "conf/myredis.conf")
	REDIS_HOST = beego.AppConfig.String("redis.host")
	REDIS_DB, _ = beego.AppConfig.Int("redis.db")
	// 建立连接池
	RedisClient = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     beego.AppConfig.DefaultInt("redis.maxidle", 1),
		MaxActive:   beego.AppConfig.DefaultInt("redis.maxactive", 10),
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			// 选择db
			c.Do("SELECT", REDIS_DB)
			return c, nil
		},
	}
}

/*连接池的使用
// 从池里获取连接
rc := RedisClient.Get()
// 用完后将连接放回连接池
defer rc.Close()
// 错误判断
if conn.Err() != nil {
  //TODO
*/