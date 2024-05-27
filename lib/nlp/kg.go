package nlp

import (
	"context"

	ollamaApi "github.com/jmorganca/ollama/api"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// note: KG stands for Knowledge Graph.
const EXTRACT_KG_SYSTEM_PROMPT = "You are a network graph maker who extracts terms and their relations from a given context. " +
	"You are provided with a context chunk (delimited by ```) Your task is to extract the ontology " +
	"of terms mentioned in the given context. These terms should represent the key concepts as per the context. \n" +
	"Thought 1: While traversing through each sentence, Think about the key terms mentioned in it.\n" +
	"\tTerms may include object, entity, location, organization, person, \n" +
	"\tcondition, acronym, documents, service, concept, etc.\n" +
	"\tTerms should be as atomistic as possible\n\n" +
	"Thought 2: Think about how these terms can have one on one relation with other terms.\n" +
	"\tTerms that are mentioned in the same sentence or the same paragraph are typically related to each other.\n" +
	"\tTerms can be related to many other terms\n\n" +
	"Thought 3: Find out the relation between each such related pair of terms. \n\n" +
	"Format your output as a list of json. Each element of the list contains a pair of terms" +
	"and the relation between them, like the following: \n" +
	"[\n" +
	"   {\n" +
	"       \"node_1\": \"A concept from extracted ontology\",\n" +
	"       \"node_2\": \"A related concept from extracted ontology\",\n" +
	"       \"edge\": \"relationship between the two concepts, node_1 and node_2 in one or two sentences\"\n" +
	"   }, {...}\n" +
	"]"

const KG_NODE_CLASS_NAME = "KGNode"
const KG_EDGE_CLASS_NAME = "KGEdge"

var KG_NODE_FIELDS = []graphql.Field{
	{Name: "name"},
}
var KG_EDGE_FIELDS = []graphql.Field{
	{Name: "name"},
	{Name: "node1"},
	{Name: "node2"},
}

type KgNode struct {
	Node1 string `json:"node_1"`
	Node2 string `json:"node_2"`
	Edge  string `json:"edge"`
}

func ExtractKnowledgeGraph(ctx context.Context, ollama *ollamaApi.Client, input string) ([]KgNode, error) {
	return ExecuteOllamaFunc[[]KgNode](ctx, ollama, EXTRACT_KG_SYSTEM_PROMPT, input)
}

// func ExportKGToVectorDB(ctx context.Context, vectordb weaviate.Client, nodes []KgNode) error {
// 	for _, n := range nodes {
// 		_, err := vectordb.Data().
// 			Creator().
// 			WithClassName("KGNode").
// 			WithProperties(map[string]string{
// 				"name": n.Node1,
// 			}).
// 			Do(ctx)
// 		if err != nil {
// 			return err
// 		}

// 		_, err = vectordb.Data().
// 			Creator().
// 			WithClassName(KG_NODE_CLASS_NAME).
// 			WithProperties(map[string]string{
// 				"name": n.Node2,
// 			}).
// 			Do(ctx)
// 		if err != nil {
// 			return err
// 		}

// 		_, err = vectordb.Data().
// 			Creator().
// 			WithClassName(KG_EDGE_CLASS_NAME).
// 			WithProperties(map[string]string{
// 				"name":  n.Edge,
// 				"node1": n.Node1,
// 				"node2": n.Node2,
// 			}).
// 			Do(ctx)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
