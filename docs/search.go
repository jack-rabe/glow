package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// todo - convert errors to logs

const (
	USER_AGENT = "go"
	SEARCH_URL = "https://proxy.search.docs.aws.amazon.com/search"
)

type SearchResponse struct {
	QueryID     string             `json:"queryId"`
	Suggestions []SearchSuggestion `json:"suggestions"`
}

type SearchSuggestion struct {
	TextExcerptSuggestion TextExcerptSuggestion `json:"textExcerptSuggestion"`
}

type TextExcerptSuggestion struct {
	Link            string        `json:"link"`
	Title           string        `json:"title"`
	SuggestionBody  string        `json:"suggestionBody"`
	Summary         string        `json:"summary"`
	Context         []ContextItem `json:"context"`
	SourceUpdatedAt int64         `json:"sourceUpdatedAt"`
	IsCitable       bool          `json:"isCitable"`
}

type ContextItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func Search(searchPhrase string) (SearchResponse, error) {
	var response SearchResponse

	requestBody := map[string]any{
		"key": "value",
		"textQuery": map[string]string{
			"input": searchPhrase,
		},
		"contextAttributes":    []map[string]string{{"key": "domain", "value": "docs.aws.amazon.com"}},
		"acceptSuggestionBody": "RawText",
		"locales":              []string{"en_us"},
	}

	// Convert body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshaling request body:", err)
		return response, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", SEARCH_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return response, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", USER_AGENT)

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 30 * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return response, err
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return response, err
	}

	return response, nil
}
