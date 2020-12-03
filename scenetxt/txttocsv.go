package scenetxt

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"scene/sceneandnode"
	"scene/scenezip"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//! 场景名称_节点名称_标准问题: SNS(scene/node/standar)
//! 场景名称_标准问题_扩展问题_答案: SSEA(scene/standar/expend/answer)
//! 场景名称_节点名称_答案: SND(scene/node/answer)

// TxtToCSV 场景文件转换成CSV文件
type TxtToCSV func(string, *csv.Writer) error

// TxtToSNS 场景文件转换成<<场景名称_节点名称_标准问题>>文件
func TxtToSNS(path string, csvWR *csv.Writer) error {
	//解析场景文件
	scene, err := SceneParse(path)
	sceneName := scene.Name
	if err != nil {
		return err
	}
	// 遍历场景中的所有节点,将有效信息写入csv文件.
	for _, node := range scene.Nodes {
		nodeName := node.NodeName
		for _, subIntent := range node.Intent {
			for _, question := range subIntent {
				//第一个问题既是标准问题
				stdQuestion := question
				if err := csvWR.Write([]string{
					sceneName,
					nodeName,
					stdQuestion,
				}); err != nil {
					return errors.Wrap(err, "txttocsv: write node info failed")
				}
				//循环执行一次即可
				break
			}

		}
	}
	return nil
}

// TxtToSSEA 场景文件转换成<<场景名称_标准问题_扩展问题_答案>>文件
func TxtToSSEA(path string, csvWR *csv.Writer) error {
	//解析场景文件
	scene, err := SceneParse(path)
	sceneName := scene.Name
	if err != nil {
		return err
	}
	// 遍历场景中的所有节点,将有效信息写入csv文件.
	for _, node := range scene.Nodes {
		//? 是入口节点吗
		if node.IsEntranceNode() {
			answer := node.Answer
			for _, subIntent := range node.Intent {

				var stdQuestion, extendQuestion string
				if len(subIntent) == 1 {
					stdQuestion = subIntent[0]
					extendQuestion = ""
					if err := csvWR.Write([]string{
						sceneName,
						stdQuestion,
						extendQuestion,
						answer,
					}); err != nil {
						return errors.Wrap(err, "txttocsv: write node info failed")
					}
					continue
				}

				for index, question := range subIntent {
					if index == 0 {
						stdQuestion = question
					} else {
						extendQuestion = question
						if err := csvWR.Write([]string{
							sceneName,
							stdQuestion,
							extendQuestion,
							answer,
						}); err != nil {
							return errors.Wrap(err, "txttocsv: write node info failed")
						}
					}
				}

			}
		}
	}
	return nil
}

// TxtToSND 场景文件转换成<<场景名称_节点名称_答案>>文件
func TxtToSND(path string, csvWR *csv.Writer) error {
	//解析场景文件
	scene, err := SceneParse(path)
	sceneName := scene.Name
	if err != nil {
		return err
	}

	// 遍历场景中的所有节点,将有效信息写入csv文件.
	for _, node := range scene.Nodes {
		//? 是全局节点吗
		if !node.IsGlobalNode() && !node.IsEntranceNode() {
			nodeName := node.NodeName
			answer := node.Answer
			if err := csvWR.Write([]string{
				sceneName,
				nodeName,
				answer,
			}); err != nil {
				return errors.Wrap(err, "txttocsv: write node info failed")
			}
		}
	}
	return nil
}

// AllTxtToCSV 将目录中所有的txt,转化成CSV文件
func AllTxtToCSV(dir string, csvName string, txtToCSVFunc TxtToCSV) error {
	//! 获取目录中所有txt文件
	txtFiles, err := scenezip.ReadFiles(dir)
	if err != nil {
		return err
	}

	//! 创建目标csv文件
	csvPath := filepath.Join(filepath.Dir(dir), csvName)
	f, err := os.Create(csvPath)
	if err != nil {
		return errors.Wrap(err, "scenetxt: create csv file failed")
	}
	csvWR := csv.NewWriter(f)
	defer csvWR.Flush()

	//! 写入列头
	fileNameOnly := strings.TrimSuffix(csvName, ".csv")

	if err := csvWR.Write(strings.Split(fileNameOnly, "_")); err != nil {
		return errors.Wrap(err, "txttocsv: write colu header failed")
	}

	//! 遍历所有文件, 将文件转换成目标csv文件
	for _, file := range txtFiles {
		err := txtToCSVFunc(file, csvWR)
		if err != nil {
			return err
		}
	}
	return nil
}

// SceneParse 根据场景txt文件构造Scene类型值
func SceneParse(sceneFile string) (*sceneandnode.Scene, error) {
	// 创建场景
	scene := sceneandnode.NewScene()
	sceneName := filepath.Base(sceneFile)
	scene.Name = strings.TrimSuffix(sceneName, ".txt")
	// 打开,并按行扫描文件
	f, err := os.Open(sceneFile)
	if err != nil {
		return nil, errors.Wrap(err, "sceneandnode: open scene file failed")
	}
	sc := bufio.NewScanner(f)
	// answerCond 话术状态
	var answerCond bool
	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimSpace(line)
		//VERSION=2.0
		if strings.HasPrefix(line, "VERSION=") {
			scene.Version = line[len("VERSION="):]
			continue
		}
		//CONFUSE_SCORE=80
		if strings.HasPrefix(line, "CONFUSE_SCORE=") {
			scene.ConfuseScore, err = strconv.Atoi(line[len("CONFUSE_SCORE="):])
			if err != nil {
				return nil, errors.Wrap(err, "txttocsv: confuse_score from str to int")
			}
			continue
		}
		//ENTRANCE_OPEN
		if strings.HasPrefix(line, "ENTRANCE_OPEN") {
			scene.EntranceOpen = true
			continue
		}
		//DOMAIN:scene564f75b86bfc4bd1ba442d9f12ff64a6
		if strings.HasPrefix(line, "DOMAIN:") {
			scene.Domain = line[len("DOMAIN:"):]
			continue
		}
		//#nodeName:民办职业培训学校终止的情形
		if strings.HasPrefix(line, "#nodeName:") {
			//关闭话术状态
			answerCond = false
			// 创建newNode
			node := sceneandnode.NewNode()
			scene.Nodes = append(scene.Nodes, node)
			node.NodeName = line[len("#nodeName:"):]
			continue
		}
		//node012555874453249327
		if strings.HasPrefix(line, "node") {
			node := scene.LastNode()
			if strings.HasSuffix(line, "#") {
				node.NodeType = sceneandnode.EntranceNode
				node.NodeID = line[len("node") : len(line)-1]
			} else if strings.HasSuffix(line, "@") {
				node.NodeType = sceneandnode.GlobalNode
				node.NodeID = line[len("node") : len(line)-1]
			} else {
				node.NodeType = sceneandnode.CommonNode
				node.NodeID = line[len("node"):]
			}
			continue
		}
		//search:city=([t_地市名称])
		if strings.HasPrefix(line, "search:") {
			node := scene.LastNode()
			node.SearchInfo = append(node.SearchInfo, line)
			continue
		}
		//e:["办理主体"]
		if strings.HasPrefix(line, "e:") {
			node := scene.LastNode()
			line = line[len("e:"):]

			var slic = make([]string, 0)
			if err := json.Unmarshal([]byte(line), &slic); err != nil {
				return nil, errors.Wrap(err, "txttocsv: parse Trigger conditions failed")
			}
			node.Intent = append(node.Intent, slic)
			continue
		}
		//a:请问您要办理的是机关事业单位基本养老保险关系跨制度转入办理流程还是职业年金跨省制度内转入办理流程？
		if strings.HasPrefix(line, "a:") {
			node := scene.LastNode()
			node.Answer = line[len("a:"):]
			answerCond = true
			continue
		}
		//childs:[045489261300023465]
		if strings.HasPrefix(line, "childs:") {
			answerCond = false // 关闭话术状态
			node := scene.LastNode()
			line = line[len("childs:"):]
			line = strings.Trim(line, "[]")
			slic := strings.Split(line, ",")
			node.Childs = append(node.Childs, slic...)
			continue
		}

		if answerCond {
			node := scene.LastNode()
			node.Answer = node.Answer + "\n" + line
			continue
		}
	}

	// 遍历场景的节点,给节点找父节点
	for _, node := range scene.Nodes {
		scene.SetFather(node)
	}

	// END
	return scene, nil
}
