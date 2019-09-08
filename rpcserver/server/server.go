package server

import (
	"github.com/lmuench/gobdb/gobdb"
)

type InsertArgs struct {
	Resource gobdb.Resource
}

type Server struct {
	DB *gobdb.DB
}

func (s *Server) Insert(args *InsertArgs, response *interface{}) error {
	return s.DB.Insert(args.Resource)
}

// func (s Server) Insert(resource interface{}, response *interface{}) error {
// 	return s.DB.Insert(resource.(gobdb.Resource))
// }
