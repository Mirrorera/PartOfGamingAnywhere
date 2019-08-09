package main

import (
	"fmt"
	"util/matchList"
	"util/myredis"
)

type ServerId = string

var N int
var M int
var mlist matchList.MatchList

func addServer() error {
	var server ServerId
	fmt.Scanln(&server)
	//fmt.Println(server)
	err := mlist.AddServer(server)
	if err != nil {
		fmt.Println("err on adding server")
		return err
	}
	fmt.Println("OK")
	return nil
}
func delServer() error {
	var server ServerId
	fmt.Scanln(&server)
	err := mlist.DeleteServer(server)
	if err != nil {
		fmt.Println("err on deleting server")
		return err
	}
	fmt.Println("OK")
	return nil
}
func getServer() error {
	server, err := mlist.GetSingleServer()
	if err != nil {
		fmt.Println("err on getting server")
		return err
	}
	fmt.Println("OK ", server)
	return nil
}
func enServerv() error {
	var server ServerId
	fmt.Scanln(&server)
	err := mlist.EnServerConn(server)
	if err != nil {
		fmt.Println("err on En server")
		return err
	}
	fmt.Println("OK")
	return nil
}
func deServer() error {
	var server ServerId
	fmt.Scanln(&server)
	err := mlist.DeServerConn(server)
	if err != nil {
		fmt.Println("err on De server")
		return err
	}
	fmt.Println("OK")
	return nil
}
func getSize() error {
	size, err := mlist.GetServerSize()
	if err != nil {
		fmt.Println("err on get size")
		return err
	}
	fmt.Println(size)
	return nil
}

//ControlDB : Any Operation
func ControlDB(oper string) error {
	var arg1, arg2 string
	fmt.Scanln(&arg1, &arg2)
	//fmt.Println(arg1, arg2)
	rc := myredis.RedisClient.Get()
	defer rc.Close()
	if arg2 == "_" {
		rep, err := rc.Do(oper, arg1)
		if err != nil {
			return err
		}
		if rep != nil {
			fmt.Println(rep)
		}
	} else {

		rep, err := rc.Do(oper, arg1, arg2)
		if err != nil {
			return err
		}
		if rep != nil {
			fmt.Println(rep)
		}
	}
	return nil
}

func main() {
	/*
	   	myredis.Init()
	   	rc := myredis.RedisClient.Get()
	   	fmt.Println("###")
	   	rc.Do("set", "myKey", "mmm")
	   	ans, err := rc.Do("get", "myKey")
	   	if err != nil {
	           fmt.Println("redis set failed:", err)
	   	}
	   	fmt.Println(ans)
	   	defer rc.Close()
	*/

	//fmt.Println("###")

	mlist.InitList("Test")
	N = 1000
	for i := 0; i < N; i++ {

		fmt.Println("#########")

		var oper string
		var err error
		fmt.Scanln(&oper)

		switch oper {
		case "Add":
			err = addServer()
		case "Del":
			err = delServer()
		case "Get":
			err = getServer()
		case "EnConn":
			err = enServerv()
		case "DeConn":
			err = deServer()
		case "GetSize":
			err = getSize()
		default:
			err = ControlDB(oper)
		}
		if err != nil {
			fmt.Println(err)
			//break
		}
	}

	return

}

/*
	myredis.Init()
	rc := myredis.RedisClient.Get()
	fmt.Println("###")
	rc.Do("set", "myKey", "mmm")
	ans, err := rc.Do("get", "myKey")
	if err != nil {
        fmt.Println("redis set failed:", err)
	}
	fmt.Println(ans)
	defer rc.Close()
*/
