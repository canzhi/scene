package sceneandnode

const (
	// EntranceNode 入口节点: 1
	EntranceNode = iota + 1
	// CommonNode 普通节点: 2
	CommonNode
	// GlobalNode 全局节点: 3
	GlobalNode
)

// Node 节点
type Node struct {
	NodeName   string     //节点名字
	NodeID     string     //节点ID,"node序列号"去掉"node""
	NodeType   int        //节点类型
	SearchInfo []string   //信息提取
	Intent     [][]string //意图
	Answer     string     //答案
	Childs     []string   //子节点集合
	Fathers    []string   //父节点集合
}

// NewNode 节点构造函数
func NewNode() *Node {
	return &Node{
		SearchInfo: make([]string, 0),
		Intent:     make([][]string, 0),
		Childs:     make([]string, 0),
		Fathers:    make([]string, 0),
	}
}

// IsValidNode 是否是有效节点
func (n *Node) IsValidNode() bool {
	return n.NodeName != ""
}

// IsEntranceNode 是否是入口节点
func (n *Node) IsEntranceNode() bool {
	return n.NodeType == EntranceNode
}

// IsCommonNode 是否是普通节点
func (n *Node) IsCommonNode() bool {
	return n.NodeType == CommonNode
}

// IsGlobalNode 是否是全局节点
func (n *Node) IsGlobalNode() bool {
	return n.NodeType == GlobalNode && len(n.Childs) == 0 && len(n.Fathers) == 0
}

// Scene 场景
type Scene struct {
	Name         string  //场景名称
	Version      string  // 版本号
	ConfuseScore int     //阈值
	EntranceOpen bool    //是否可重新返回入口
	Domain       string  //领域ID
	Nodes        []*Node // 节点集合
}

// NewScene 构造场景
func NewScene() *Scene {
	// 场景初始化
	return &Scene{
		Nodes: make([]*Node, 0),
	}
}

// Length 场景中元素个数
func (s *Scene) Length() int {
	return len(s.Nodes)
}

// LastNode 获取最后一个元素
func (s *Scene) LastNode() *Node {
	return s.Nodes[s.Length()-1]
}

// SetFather 给节点配置父节点
func (s *Scene) SetFather(n *Node) {
	for _, node := range s.Nodes {
		for _, nodeID := range node.Childs {
			if n.NodeID == nodeID {
				n.Fathers = append(n.Fathers, node.NodeID)
			}
		}
	}
}
