package executors

import (
	"bioflows/models/pipelines"
	"bioflows/resolver"
	"github.com/goombaio/dag"
)

type DagScheduler struct {
	rankedNodes map[string] int
	nodeNames map[string] *dag.Vertex
	pipeline *pipelines.BioPipeline
}
func (d *DagScheduler) getValue(node *dag.Vertex) int {
	if val , ok := d.rankedNodes[resolver.ResolveToolKey(node.ID,d.pipeline.ID)]; ok {
		return val
	}
	return 0
}

func (d *DagScheduler) getMaxRank() int {
	max := -1
	for _ , val := range d.rankedNodes {
		if val > max {
			max = val
		}
	}
	return max
}

func (d *DagScheduler) prepareNodes() ([][]*dag.Vertex,error) {
	sortedNodes := make([][]*dag.Vertex,d.getMaxRank()+1)
	for key , val := range d.rankedNodes {
		if sortedNodes[val] == nil {
			sortedNodes[val] = make([]*dag.Vertex,1)
		}
		sortedNodes[val] = append(sortedNodes[val],d.nodeNames[key])
	}
	return sortedNodes , nil
}
func (d *DagScheduler) Rank(parentPipeline *pipelines.BioPipeline , pb *dag.DAG) ([][]*dag.Vertex,error){
	d.pipeline = parentPipeline
	d.rankedNodes = make(map[string]int)
	d.nodeNames = make(map[string]*dag.Vertex)
	parents := pb.SourceVertices()
	rank := 0
	for _ , parent := range parents {
		d.rankVertex(rank, parent)
	}
	preparedList , err := d.prepareNodes()
	if err != nil {
		return nil , err
	}
	rankedList := make([][]*dag.Vertex,0)
	total := len(preparedList)
	for idx := 0 ; idx < total;idx++ {
		sublist := preparedList[idx]
		if sublist == nil || len(sublist) == 0{
			continue
		}
		rankedList = append(rankedList,sublist)
	}
	return rankedList , nil

}
func (d *DagScheduler) rankVertex(pRank int, node *dag.Vertex) {
	currentNodeName := resolver.ResolveToolKey(node.ID,d.pipeline.ID)
	val := d.getValue(node)
	nodeRank := val + pRank +  node.InDegree()
	if _ , ok := d.rankedNodes[currentNodeName] ; !ok {
		d.rankedNodes[currentNodeName] = nodeRank
	}
	d.nodeNames[resolver.ResolveToolKey(node.ID,d.pipeline.ID)] = node
	for _ , child := range node.Children.Values() {
		childNode := child.(*dag.Vertex)
		d.rankVertex(nodeRank,childNode)
	}
}
