package nlp

import (
	"context"

	ollamaApi "github.com/jmorganca/ollama/api"
)

// note: NER stands for Named Entity Recognition
const EXTRACT_NER_SYSTEM_PROMPT = "Your task is to extract the key entities mentioned in the users input.\n" +
	"Entities may include - event, concept, person, place, object, document, organization, artifact, misc, etc.\n" +
	"Format your output as a list of json with the following structure.\n" +
	"[{\n" +
	"   \"entity\": The Entity string\n" +
	"   \"importance\": How important is the entity given the context on a scale of 1 to 5, 5 being the highest.\n" +
	"   \"type\": Type of entity\n" +
	"}, { }]"

type NerEntity struct {
	Entity     string `json:"entity"`
	Importance int    `json:"importance"`
	Type       string `json:"type"`
}

func ExtractNamedEntities(ctx context.Context, ollama *ollamaApi.Client, input string) ([]NerEntity, error) {
	return ExecuteOllamaFunc[[]NerEntity](ctx, ollama, EXTRACT_NER_SYSTEM_PROMPT, input)
}
