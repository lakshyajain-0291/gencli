package gemini

import (
	"context"
	"fmt"
	"gemini_cli_tool/fileinfo"
	"math"

	"github.com/google/generative-ai-go/genai"
)

func SearchRelevantFiles(files []fileinfo.FileInfo, query string, relevanceIndex float32) (int, error) {
	ctx := context.Background()

	setupSession, err := NewsetupSession(ctx)
	if err != nil {
		return -1, err
	}

	em := setupSession.client.EmbeddingModel("text-embedding-004")
	res, err := em.EmbedContent(ctx, genai.Text(query))
	if err != nil {
		return -1, err
	}

	queryEmbedding := res.Embedding.Values

	// var results []int
	var result int = -1

	var maxSimilarity float32 = 0.0
	for i, file := range files {
		similarity := cosineSimilarity(file.Embedding, queryEmbedding)

		fmt.Printf("\n||Similarity With %s : %f||\n", file.Name, similarity)

		if similarity > maxSimilarity && file.Description != "" {
			maxSimilarity = similarity
			result = i
		}
		// if similarity > relevanceIndex {
		// 	results = append(results, i)
		// }
	}

	// return results, nil
	if maxSimilarity > 0.35 {
		return result, nil
	}

	return -1, nil
}

// Error handling for cosine similarity.---olama, lamaindex ,external packages
func cosineSimilarity(vec1 []float32, vec2 []float32) float32 {
	var dotProduct, normVec1, normVec2 float32

	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		normVec1 += vec1[i] * vec1[i]
		normVec2 += vec2[i] * vec2[i]
	}

	return dotProduct / (float32(math.Sqrt(float64(normVec1))) * float32(math.Sqrt(float64(normVec2))))
}
