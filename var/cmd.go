package _var

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

var (
	/*	CmdMode		= xsf.Mode				// -m
		CmdCfg		= xsf.Cfg				// -c
		CmdProject	= xsf.Project			// -p
		CmdGroup	= xsf.Group				// -g
		CmdService	= xsf.Service			// -s
		CmdCompanionUrl = xsf.CompanionUrl	// -u
	*/
	// default 缺省配置模式为native
	CmdCfg       = flag.String("f", "xtest.toml", "client cfg name")
	XTestVersion = flag.String("v", "2.5.2", "xtest version")
)

func Usage() {
	fmt.Println("usage of common test tool")
	fmt.Println("-f		specify config file")
	os.Exit(0)
}

// Input jbzhou5 Input data
func Input(data string) (int, error) {
	in := bufio.NewReader(os.Stdin)
	fmt.Print("Please input data: ")
	n, err := fmt.Fscanln(in, &data)
	if err != nil {
		return 0, err
	}
	return n, nil
}
