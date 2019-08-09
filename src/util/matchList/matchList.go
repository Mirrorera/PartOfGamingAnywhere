package matchList

import (
	"errors"
	"util/myredis"

	"github.com/garyburd/redigo/redis"
)

//单次请求获取最大服务器数量
const SINGLE_REQUEST_SERVERS_NUM = 10

//需要额外编写Server 这里为方便编译
//type Server struct{
//	Name string
//}

type ServerId = string

//type ServerList = queue.Queue

type MatchList struct {
	//availServerList ServerList
	//true表示使用中 false表示服务器不可用
	//
	//inavailServerList map[*Server]bool
	//inavailServerList redis.Conn
	//RedisConn redis.Conn

	availList  string //List 实质Deque
	usingSet   string //Set
	unavailSet string //Set
	availSet   string // Set

	gameType string
	//serverSize int
	//serverCap int
}

/*
func (this *MatchList)getServerList() []*Server {
	return this.enServerList
}
*/

func (this *MatchList) InitList(GameName string) {
	this.gameType = GameName
	//需预先执行myredis.Init()操作
	//rc = myredis.RedisClient.Get()
	myredis.Init()
	this.availList = "availSL"
	this.usingSet = "usingSL"
	this.unavailSet = "unavailSL"
	this.availSet = "availSet"
	rc := myredis.RedisClient.Get()
	rc.Do("sadd", this.usingSet, "####")
	rc.Do("sadd", this.unavailSet, "#####")
	rc.Do("sadd", this.availSet, "####@##")

	//emm, err := redis.Int(rc.Do("sismember", this.availSet, "####@##"))
	//fmt.Println(emm, err)
	defer rc.Close()
}

//获取列表种类
func (this *MatchList) SetGameType(GameName string) {
	this.gameType = GameName
}
func (this *MatchList) GetGameType() string {
	return this.gameType
}

//获取列表大小
func (this *MatchList) GetServerSize() (int, error) {
	rc := myredis.RedisClient.Get()
	defer rc.Close()
	//比实际可用服务器多一
	//fmt.Println(this.availSet == this.availSet)
	//fmt.Println(this.availSet)
	//this.availSet = this.availSet
	//fmt.Println(this.availSet)
	size, err := redis.Int(rc.Do("scard", this.availSet))
	size--
	if err != nil {
		return -1, err
	}
	return size, nil
}

/*
func (this *MatchList)CloseConn() {
	defer rc.Close()
}
*/
//获取单个服务器
func (this *MatchList) GetSingleServer() (server string, err error) {
	var exist int
	rc := myredis.RedisClient.Get()
	defer rc.Close()
	for true {
		//fmt.Println("---")
		server, err = redis.String(rc.Do("lpop", this.availList))
		if err != nil {
			return "", err
		}

		exist, err = redis.Int(rc.Do("sismember", this.availSet, server))
		if err != nil {
			return "", err
		}
		if exist != 1 {
			continue
		}

		break
	}
	//fmt.Println("#", server)
	rc.Do("rpush", this.availList, server)
	return
}

//获取多个服务器
func (this *MatchList) GetServers() (serverlist []ServerId, num int, err error) {
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

//增加可用服务器
func (this *MatchList) AddServer(server ServerId) error {
	rc := myredis.RedisClient.Get()
	defer rc.Close()
	//判断服务器是否存在且可用
	exist, err := redis.Int(rc.Do("sismember", this.availSet, server))
	if err != nil {
		return err
	}
	if exist == 1 {
		return errors.New("Server hash been existed")
	}
	//判断服务器是否处于停用中
	exist, err = redis.Int(rc.Do("sismember", this.unavailSet, server))
	if err != nil {
		//fmt.Println("is existed error")
		return err
	}
	if exist == 1 {
		_, err = redis.Int(rc.Do("srem", this.unavailSet, server))
		if err != nil {
			//fmt.Println("is existed2 error")
			return err
		}
	}
	//添加至可用列表
	_, err = redis.Int(rc.Do("rpush", this.availList, server))
	if err != nil {
		return err
	}
	_, err = redis.Int(rc.Do("sadd", this.availSet, server))
	if err != nil {
		return err
	}
	return nil
}

//删除服务器，移入不可用列表并标注不可用
func (this *MatchList) DeleteServer(server ServerId) error {
	rc := myredis.RedisClient.Get()
	defer rc.Close()
	//T判断服务器是否存在
	exist, err := redis.Int(rc.Do("sismember", this.availSet, server))
	if err != nil {
		return err
	}
	if exist != 1 {
		return errors.New("The server does not exist")
	}

	_, err = rc.Do("srem", this.availSet, server)
	if err != nil {
		return err
	}
	_, err = rc.Do("sadd", this.unavailSet, server)
	if err != nil {
		return err
	}
	return nil
}

//创建服务器连接，移入不可用列表并标注正忙
func (this *MatchList) EnServerConn(server ServerId) error {
	rc := myredis.RedisClient.Get()
	defer rc.Close()

	//判断服务器是否存在
	exist, err := redis.Int(rc.Do("sismember", this.availSet, server))
	if err != nil {
		return err
	}
	if exist != 1 {
		return errors.New("The server does not exist")
	}

	_, err = rc.Do("srem", this.availSet, server)
	if err != nil {
		return err
	}
	_, err = rc.Do("sadd", this.usingSet, server)
	if err != nil {
		return err
	}
	return nil
}

//关闭服务器连接，移入可用列表
func (this *MatchList) DeServerConn(server ServerId) error {
	rc := myredis.RedisClient.Get()
	defer rc.Close()

	//判断是否存在于usingList
	exist, err := redis.Int(rc.Do("sismember", this.usingSet, server))
	if err != nil {
		return nil
	}
	if exist != 1 {
		err = errors.New("No Server to be Closed")
		return err
	}
	_, err = rc.Do("srem", this.usingSet, server)
	if err != nil {
		return err
	}

	//判断服务器是否已经可用
	exist, err = redis.Int(rc.Do("sismember", this.availSet, server))
	if err != nil {
		return err
	}
	if exist == 1 {
		return errors.New("The server has been available")
	}
	_, err = rc.Do("sadd", this.availSet, server)
	if err != nil {
		return err
	}
	_, err = rc.Do("lpush", this.availList, server)
	if err != nil {
		return err
	}
	return err
}
