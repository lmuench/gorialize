package main

import (
	"encoding/gob"
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/lmuench/gobdb/gobdb"
	"github.com/lmuench/gobdb/rpcexample/model"
	"github.com/lmuench/gobdb/rpcserver/server"
)

func main() {
	gob.Register(&model.User{})

	// Server
	srv := &server.Server{
		DB: gobdb.NewDB(
			"/tmp/gobdb/rpcserver/example_dev",
			true,
		),
	}

	_ = rpc.Register(srv)
	rpc.HandleHTTP()
	lstn, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go http.Serve(lstn, nil)

	// Client
	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	user := &model.User{
		Name: "John Doe",
		Age:  42,
	}

	var response string

	err = client.Call("Server.Insert", &server.InsertArgs{Resource: user}, &response)
	if err != nil {
		log.Fatal("server error:", err)
	}
}
