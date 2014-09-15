package main

import (
	"cfm"
	"log"
)

// root
type RootConf struct {
	Test  int
	Test2 string
	Test3 []int
}

var RootCommands = []cfm.Command{
	{"test", "Test", 1, cfm.CommandSetInt},
	{"test2", "Test2", "wxw", cfm.CommandSetString},
	{"test3", "Test3", "wxw", cfm.CommandSetIntArray},
}

// tcp
type TcpConf struct {
	Tcp int
}

var TcpCommands = []cfm.Command{
	{"tcp", "Tcp", 111, cfm.CommandSetInt},
}

func main() {
	cfg := cfm.LoadConfig("test.conf")

	rootConf := &RootConf{}
	root := cfm.NewRootContext(cfg)
	root.AddCommand(RootCommands)
	root.AddConf(rootConf)

	tcpConf := &TcpConf{}
	tcp, err := root.AddContext("tcp")
	if err != nil {
		log.Println(err)
	}
	tcp.AddCommand(TcpCommands)
	tcp.AddConf(tcpConf)

	if err := cfg.Parse(); err != nil {
		log.Fatalln(err)
	}

	log.Println(rootConf, tcpConf)
}
