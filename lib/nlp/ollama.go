package nlp

import (
	"bytes"
	"context"
	"encoding/json"

	ollamaApi "github.com/jmorganca/ollama/api"
)

func ExecuteOllamaFunc[T any](
	ctx context.Context,
	ollama *ollamaApi.Client,
	system string,
	input string,
) (T, error) {
	var result T

	buff := bytes.NewBuffer(nil)
	err := ollama.Generate(
		ctx,
		&ollamaApi.GenerateRequest{
			Model:  "zephyr",
			System: system,
			Prompt: input,
		},
		func(gr ollamaApi.GenerateResponse) error {
			_, err := buff.WriteString(gr.Response)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(buff.Bytes(), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
