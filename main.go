package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scene/scenetxt"
	"scene/scenezip"

	"github.com/pkg/errors"
)

var src, dst string

//! 命令行参数
var (
	s string //输入源目录
	o string //输出csv模式
	h bool   //帮助
)

func init() {

	//! 命令行参数
	flag.StringVar(&s, "s", "src", "set `directory` which the source scene file from to the aim of csv file")
	flag.StringVar(&o, "o", "", "set `csvFileType` then output the csvFile:\n\n  sns: 场景名称_节点名称_标准问题(scene/node/standar)\n  ssea: 场景名称_标准问题_扩展问题_答案(scene/standar/expend/answer)\n  snd: 场景名称_节点名称_答案(scene/node/answer)\n")
	flag.BoolVar(&h, "h", false, "this help")
	flag.Usage = usage
	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(1)
	}

	//当前路径
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// 初始化路径
	src = filepath.Join(pwd, s)
	dst = filepath.Join(pwd, "dst")
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `scene version: scene/1.0.0
Usage: scene [-h] [-o csvFileType]
Options:
`)
	flag.PrintDefaults()
}

func main() {

	//! 解压缩场景文件(zip)为场景文件(txt)
	if err := scenezip.ExtractFromZipToTxt(src, dst); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	//! 根据命令行参数,输出不同的CSV文件
	switch o {
	case "sns":
		if err := scenetxt.AllTxtToCSV(dst, "场景名称_节点名称_标准问题.csv", scenetxt.TxtToSNS); err != nil {
			fmt.Println(errors.Cause(err))
			os.Exit(1)
		}
	case "ssea":
		if err := scenetxt.AllTxtToCSV(dst, "场景名称_标准问题_扩展问题_答案.csv", scenetxt.TxtToSSEA); err != nil {
			fmt.Println(errors.Cause(err))
			os.Exit(1)
		}
	case "snd":
		if err := scenetxt.AllTxtToCSV(dst, "场景名称_节点名称_答案.csv", scenetxt.TxtToSND); err != nil {
			fmt.Println(errors.Cause(err))
			os.Exit(1)
		}
	default:
		fmt.Println("please input:sns/ssea/snd,you can use -h to see the usage.")
	}
}
