package pipelines

import (
	"bytes"
	"fmt"
	viz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/goombaio/dag"
	"strings"
)



func PreparePipeline(b *BioPipeline,funcCall func (b *BioPipeline) *BioPipeline) (*BioPipeline,error) {
	//TODO: this function should perform the following tasks
	// 1. Download the tool from the remote repository, in this order (URL , Bioflows Hub)
	// 2. Update the downloaded tool parameters by the newly written parameters.
	return b , nil
}

func GraphContains(g *dag.DAG, v string) bool {
	_ , err := g.GetVertex(v)
	return err == nil
}

func CreateGraph(b *BioPipeline) (*dag.DAG,error){
	g := dag.NewDAG()
	processedSteps := make(map[string]*dag.Vertex)
	for _ , step := range b.Steps {
		step.Prepare()
		if len(step.Depends) <= 0{
			vertex := dag.NewVertex(step.ID,step)
			g.AddVertex(vertex)
			processedSteps[step.ID] = vertex
		}else{
			from := step.Depends
			fromNodes := strings.Split(from,",")
			for _ , fromNode := range fromNodes{
				currentVertex := dag.NewVertex(step.ID,step)
				if GraphContains(g,step.ID){
					currentVertex , _ = g.GetVertex(step.ID)
				}
				if parentVertex, ok := processedSteps[fromNode]; !ok {
					panic(fmt.Errorf("Unknown Bioflows Step mentioned in %s",step.Name))
				}else{
					g.AddVertex(currentVertex)
					g.AddEdge(parentVertex,currentVertex)

					processedSteps[step.ID] = currentVertex
				}
			}
		}
	}
	return g, nil
}

func ToDotGraph(b *BioPipeline, d *dag.DAG) (string,error){
	parents := d.SourceVertices()
	g := viz.New()
	graph , err := g.Graph()
	graph.SetLabel(b.Name)

	if err != nil {
		return "",err
	}
	defer graph.Close()
	defer g.Close()
	for _ , parent := range parents {
		current := parent.Value.(BioPipeline)
		parentNode , _ :=graph.CreateNode(current.Name)
		if parent.Children.Size() > 0 {
			for _ , child := range parent.Children.Values(){
				currentChild := (child.(*dag.Vertex)).Value.(BioPipeline)
				currentChildNode , _ := graph.CreateNode(currentChild.Name)
				edgeName := fmt.Sprintf("%s To %s",current.Name,currentChild.Name)
				graph.CreateEdge(edgeName,parentNode,currentChildNode)
				//edge.SetLabel(edgeName)
				appendChildren(graph,child.(*dag.Vertex),currentChildNode)
			}
		}
	}
	var buff bytes.Buffer
	if err = g.Render(graph,"dot",&buff); err != nil{
		return "" , err
	}
	return buff.String(), nil
}
func appendChildren(graph *cgraph.Graph,current *dag.Vertex, currentNode *cgraph.Node){
	if current.Children.Size() <= 0{
		return
	}else{
		currentPipeline := current.Value.(BioPipeline)
		for _ , child := range current.Children.Values(){
			currentChild := (child.(*dag.Vertex)).Value.(BioPipeline)
			ChildNode, _ := graph.CreateNode(currentChild.Name)
			edgeName := fmt.Sprintf("%s To %s", currentPipeline.Name,currentChild.Name)
			graph.CreateEdge(edgeName,currentNode, ChildNode)
			//edge.SetLabel(edgeName)
			appendChildren(graph,child.(*dag.Vertex),ChildNode)
		}
	}
}

