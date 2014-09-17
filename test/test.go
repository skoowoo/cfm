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
	{"test3", "Test3", "1 2 3", cfm.CommandSetIntArray},
}

// tcp
type TcpConf struct {
	Tcp int
}

var TcpCommands = []cfm.Command{
	{"tcp", "Tcp", 111, cfm.CommandSetInt},
}

var (
	rootConf *RootConf
	tcpConf  *TcpConf
)

func main() {
	if err := cfm.LoadConfig("test.conf").Parse(); err != nil {
		log.Fatalln(err)
	}

	log.Println(rootConf, tcpConf)
}

func init() {
	rootConf = &RootConf{}
	root := cfm.NewRootContext()
	root.AddCommand(RootCommands)
	root.AddConf(rootConf)

	tcpConf = &TcpConf{}
	tcp := cfm.NewContext("tcp", cfm.CTX_ROOT_NAME)
	tcp.AddCommand(TcpCommands)
	tcp.AddConf(tcpConf)
}
