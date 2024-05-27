package main

import (
	"context"
	"log"
	"research-toolkit/lib/nlp"

	ollamaApi "github.com/jmorganca/ollama/api"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

func generateKg(ctx context.Context, ollama *ollamaApi.Client, vectordb *weaviate.Client) error {
	nodes, err := nlp.ExtractKnowledgeGraph(
		context.Background(),
		ollama,
		"The polar bear (Ursus maritimus) is a large bear native to the Arctic and nearby areas. It is closely related to the brown bear, and the two species can interbreed. The polar bear is the largest extant species of bear and land carnivore, with adult males weighing 300–800 kg (660–1,760 lb). The species is sexually dimorphic, as adult females are much smaller. The polar bear is white- or yellowish-furred with black skin and a thick layer of fat. It is more slender than the brown bear, with a narrower skull, longer neck and lower shoulder hump. Its teeth are sharper and more adapted to cutting meat. The paws are large and allow the bear to walk on ice and paddle in the water.",
	)
	if err != nil {
		return err
	}

	err = nlp.ExportKGToVectorDB(
		context.Background(),
		*vectordb,
		nodes,
	)
	return err
}

func ask(ctx context.Context, ollama *ollamaApi.Client, db *weaviate.Client, question string) (*models.GraphQLResponse, error) {
	result, err := db.GraphQL().
		Get().
		WithClassName(nlp.KG_NODE_CLASS_NAME).
		WithFields(nlp.KG_NODE_FIELDS...).
		WithLimit(100).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	ollama, err := ollamaApi.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	vectordb, err := weaviate.NewClient(weaviate.Config{
		Host:   "localhost:8080",
		Scheme: "http",
	})
	if err != nil {
		log.Fatal(err)
	}

	res, err := ask(context.Background(), ollama, vectordb, "polar bear")
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range res.Data {
		log.Println("KEY", k, "VALUE", v)
	}
}
