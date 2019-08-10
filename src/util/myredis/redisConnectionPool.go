package myredis

import (
	"log"
	//"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/ini.v1"
)

var (
	// 定义常量
	RedisClient *redis.Pool
	//REDIS_HOST string
	//REDIS_DB   int
)

type Config struct {
	REDIS_HOST      string `ini:"redis.host"`
	REDIS_DB        int    `int:"redis.db"`
	REDIS_MAXIDLE   int    `ini:"redis.maxidle"`
	REDIS_MAXACTIVE int    `int:"redis.maxactive"`
}

//var filepath = "C:/Users/sunyu/Desktop/vsdata/Go/Dev/G1/Tracker/conf/myredis.conf"
//var filepath = "./conf/myredis.conf"

func ReadConfig(filepath string) (Config, error) {
	var config Config
	conf, err := ini.Load(filepath)
	if err != nil {
		log.Println("load config file fail!")
		return config, err
	}
	conf.BlockMode = false
	err = conf.MapTo(&config)
	if err != nil {
		log.Println("mapto config file fail!")
		return config, err
	}
	return config, err
}
func Init(filepath string) {

	// 从配置文件获取redis的ip以及db
	config, err := ReadConfig(filepath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(config)

	// 建立连接池
	RedisClient = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive
		MaxIdle:     config.REDIS_MAXIDLE,   //beego.AppConfig.DefaultInt("redis.maxidle", 1),//
		MaxActive:   config.REDIS_MAXACTIVE, //beego.AppConfig.DefaultInt("redis.maxactive", 10),
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.REDIS_HOST)
			if err != nil {
				return nil, err
			}
			// 选择db
			c.Do("SELECT", config.REDIS_DB)
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
