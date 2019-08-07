package matchList

import (
	//"fmt"
	"util/myredis"
	"github.com/garyburd/redigo/redis"
	"errors"
)

//单次请求获取最大服务器数量
const SINGLE_REQUEST_SERVERS_NUM = 10

//需要额外编写Server 这里为方便编译
//type Server struct{
//	Name string
//}

type ServerId = string

//type ServerList = queue.Queue

type MatchList struct{
	//availServerList ServerList
	//true表示使用中 false表示服务器不可用
	//
	//inavailServerList map[*Server]bool 
	//inavailServerList redis.Conn
	//RedisConn redis.Conn

	//availServerList interface{}//List 实质Deque
	//usingServerList interface{}//Set
	//unavailServerList interface{}//Set

	gameType string
	serverSize int
	//serverCap int
}

/*
func (this MatchList)getServerList() []*Server {
	return this.enServerList
}
*/

func (this MatchList)InitList() {
	//需预先执行myredis.Init()操作
	//rc = myredis.RedisClient.Get()
	myredis.Init()
	//this.availServerList = "availSL"
	//this.usingServerList = "usingSL"
	//this.unavailServerList = "unavailSL"
	rc := myredis.RedisClient.Get()
	rc.Do("sadd", "usingSL", "#####")
	rc.Do("sadd", "unavailSL", "#####")
	defer rc.Close()
}
/*
func (this MatchList)CloseConn() {
	defer rc.Close()
}
*/
//获取单个服务器
func (this MatchList)GetSingleServer() (server string, err error) {
	rc := myredis.RedisClient.Get()
	defer rc.Close() 
	for true {
		//fmt.Println("---")
		server, err = redis.String(rc.Do("lpop", "availSL"))
		//fmt.Println(server)
		//server = t_server
		//fmt.Println("#", redis.ErrNil)
		if err != nil {
			//fmt.Println(server, err)
			//fmt.Println("ttt", err)
			return "", err
		}
		exist, err := redis.Int(rc.Do("sismember", "usingSL", server))
		if err != nil || exist == 1 {
			//fmt.Println("P", err)
			continue
		}
		exist, err = redis.Int(rc.Do("sismember", "unavailSL", server))
		if err != nil || exist == 1 {
			//fmt.Println("Q")
			continue
		}
		break
	}
	//fmt.Println("#", server)
	rc.Do("rpush", "availSL", server)
	return
}

//获取多个服务器

func (this MatchList)GetServers() (serverlist []ServerId, num int, err error){
	loop := SINGLE_REQUEST_SERVERS_NUM
	for loop > 0 {
		loop--
		server, err := this.GetSingleServer()
		if err != nil {
			return serverlist, num, err
		}
		serverlist = append(serverlist, server)
		num++
	}
	return
}

//获取列表种类
func (this MatchList)GetGameType() string {
	return this.gameType
}

//获取列表大小
func (this MatchList)GetServerSize() int {
	return this.serverSize
}

//增加可用服务器
func (this MatchList)AddServer(server ServerId) error {
	rc := myredis.RedisClient.Get()
	defer rc.Close() 
	//判断服务器是否存在
	exist, err := redis.Int(rc.Do("sismember", "unavailSL", server))
	if err != nil {
		//fmt.Println("is existed error")
		return err
	}
	if exist == 1 {
		_, err = redis.Int(rc.Do("srem", "unavailSL", server))
		if err != nil {
			//fmt.Println("is existed2 error")
			return err
		}
	}
	_, err = redis.Int(rc.Do("rpush", "availSL", server))
	return err
}

//删除服务器，移入不可用列表并标注不可用
func (this MatchList)DeleteServer(server ServerId) error{
	rc := myredis.RedisClient.Get()
	defer rc.Close()
	//Todo 未判断服务器是否存在，需要添加一Set记录可用服务器存在性
	_, err := rc.Do("sadd", "unavailSL", server)
	return err
}

//创建服务器连接，移入不可用列表并标注正忙
func (this MatchList)EnServerConn(server ServerId) error{
	rc := myredis.RedisClient.Get()
	defer rc.Close() 
	//Todo 未判断服务器是否存在，需要添加一Set记录可用服务器存在性
	_, err := rc.Do("sadd", "usingSL", server)
	return err
}

//关闭服务器连接，移入可用列表
func (this MatchList)DeServerConn(server ServerId) error{
	rc := myredis.RedisClient.Get()
	defer rc.Close() 
	exist, err := redis.Int(rc.Do("sismember", "usingSL", server))
	if err != nil {
		return nil
	}
	if exist == 0 {
		err = errors.New("No Server to be Closed")
		return err
	}
	_, err = rc.Do("srem", "usingSL", server)
	if err != nil {
		return err
	}
	_, err = rc.Do("lpush", "availSL", server)
	return err
}
