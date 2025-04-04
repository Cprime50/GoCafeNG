package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	geminiEndpoint = "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-pro:generateContent"
	maxRetries     = 3
	retryDelay     = 2 * time.Second
)

// JobInfo contains the minimal job information needed for description enhancement
type JobInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Description string `json:"description"`
}

// GeminiRequest represents the request structure for the Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content represents the content part of the Gemini request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a part in the content
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from the Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

// JobDescriptionHTML represents the structure for the enhanced job description
type JobDescriptionHTML struct {
	JobID       string `json:"job_id"`
	HTMLContent string `json:"html_content"`
}

// EnhanceJobDescriptions takes job information and enhances their descriptions using Gemini API
func EnhanceJobDescriptions(jobsInfo []JobInfo) (map[string]string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not found in environment variables")
	}

	// Convert job data to JSON for the prompt
	jobsJSON, err := json.Marshal(jobsInfo)
	if err != nil {
		return nil, fmt.Errorf("error marshaling jobs data: %w", err)
	}

	// Create the prompt for Gemini
	prompt := fmt.Sprintf(`
You are a job description formatter for a job board website. 
I will provide you with a list of job descriptions in JSON format. 
Your task is to format each job description into clean, well-structured HTML that will be displayed on a job board website.

For each job description:
1. Maintain the original content but improve formatting with proper HTML tags.
2. Use semantic HTML elements (<h1>, <h2>, <h3>, <p>, <ul>, <li>, etc.) appropriately.
3. Highlight important sections like "Requirements", "Responsibilities", "Benefits", etc.
4. Make the content more readable with proper spacing and organization.
5. Do not add any information that is not in the original description.
6. Do not remove any information from the original description.
7. Return the results as a JSON array of objects with the following structure:
[
  {
    "job_id": "the original job ID",
    "html_content": "the formatted HTML content"
  },
  ...
]

Here are the job descriptions to format:
%s
`, string(jobsJSON))

	// Create the Gemini API request
	request := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: prompt,
					},
				},
			},
		},
	}

	// Convert request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Call the Gemini API with retries
	var responseBody []byte
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		responseBody, err = callGeminiAPI(ctx, apiKey, requestJSON)
		if err == nil {
			break
		}

		log.Printf("Attempt %d failed: %v", attempt, err)
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("all attempts to call Gemini API failed: %w", err)
	}

	// Parse the response
	var response GeminiResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Check if the response was blocked
	if response.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("gemini API blocked the request: %s", response.PromptFeedback.BlockReason)
	}

	// Check if we have candidates
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in Gemini API response")
	}

	// Extract the formatted HTML content
	formattedContent := response.Candidates[0].Content.Parts[0].Text

	// Parse the JSON response from Gemini
	var enhancedDescriptions []JobDescriptionHTML
	
	// Try to extract JSON from the response text (it might be surrounded by markdown code blocks)
	jsonContent := extractJSON(formattedContent)
	
	if err := json.Unmarshal([]byte(jsonContent), &enhancedDescriptions); err != nil {
		return nil, fmt.Errorf("error unmarshaling enhanced descriptions: %w", err)
	}

	// Create a map of job ID to HTML content
	result := make(map[string]string)
	for _, desc := range enhancedDescriptions {
		result[desc.JobID] = desc.HTMLContent
	}

	return result, nil
}

// extractJSON tries to extract JSON content from a string that might contain markdown or other text
func extractJSON(content string) string {
	// Check if content is wrapped in markdown code blocks
	jsonStart := 0
	jsonEnd := len(content)

	// Look for JSON array start after removing possible markdown code blocks
	if start := bytes.Index([]byte(content), []byte("[")); start != -1 {
		jsonStart = start
	}

	// Look for JSON array end
	if end := bytes.LastIndex([]byte(content), []byte("]")); end != -1 {
		jsonEnd = end + 1
	}

	// Extract the JSON part
	if jsonStart < jsonEnd {
		return content[jsonStart:jsonEnd]
	}

	return content
}

// callGeminiAPI makes the actual HTTP request to the Gemini API
func callGeminiAPI(ctx context.Context, apiKey string, requestJSON []byte) ([]byte, error) {
	// Create the URL with API key
	url := fmt.Sprintf("%s?key=%s", geminiEndpoint, apiKey)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
