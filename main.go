package main

import (
	"flag"
	"log"
	"xtest/conf"

	xsfcli "github.com/xfyun/xsf/client"
	"github.com/xfyun/xsf/utils"
)

func main() {
	var (
		f string
		v bool
	)

	flag.StringVar(&f, "f", "xtest.toml", "client cfg name")
	flag.BoolVar(&v, "v", false, "show xtest version")
	flag.Parse()

	if v {
		log.Println("3.0.0")
		return
	}

	cli, err := xsfcli.InitClient(
		conf.CliName,
		utils.Native,
		utils.WithCfgName(f),
	)
	if err != nil {
		log.Fatalf("cli xsf init failed: %v", err)
	}

	// cli配置初始化;
	c := conf.NewConf()
	if err = c.ConfInit(cli.Cfg()); err != nil {
		log.Fatalf("cli conf init failed:  %v", err)
	}

	x := NewXtest(cli, c)
	x.Run()
}
