package vdb

import (
	"research-toolkit/lib/nlp"

	chroma_go "github.com/amikos-tech/chroma-go"
)

type VectorDB struct {
	chroma *chroma_go.Client
}

func NewVectorDB(url string) VectorDB {
	client := chroma_go.NewClient(url)
	return VectorDB{
		chroma: client,
	}
}

func (db VectorDB) StoreKG(graph []nlp.KgNode) error {
	nodes, err := db.chroma.CreateCollection(
		"node",
		map[string]interface{}{},
		true,
		nil,
		chroma_go.L2,
	)
	if err != nil {
		return err
	}

	for _, rel := range graph {
		// embeddings, err := nodes.EmbeddingFunction.CreateEmbedding([]string{
		// 	rel.Node1,
		// })
		// if err != nil {
		// 	return err
		// }
		nodes.Add(
			nil,
			[]map[string]any{},
			[]string{},
			[]string{},
		)
	}
}
