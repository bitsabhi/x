package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"newsapp/models"
	"os"
)

const (
	defaultNewsAPIKey   = "your_default_newsapi_key"
	defaultOpenAIAPIKey = "your_default_openai_api_key"
)

// Struct for OpenAI Request
type OpenAIRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// Struct for OpenAI Response
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func getAPIKeys() (string, string) {
	newsAPIKey := os.Getenv("NEWSAPI_KEY")
	if newsAPIKey == "" {
		newsAPIKey = defaultNewsAPIKey
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		openaiAPIKey = defaultOpenAIAPIKey
	}
	return newsAPIKey, openaiAPIKey
}

func FetchAndStoreNewsOpenAI(db *gorm.DB) {
	if db == nil {
		log.Fatalf("Database connection is nil")
		return
	}

	newsAPIKey, _ := getAPIKeys()

	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=us&category=business&apiKey=%s", newsAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching news: %v", err)
		return
	}
	defer resp.Body.Close()

	// Log the status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: received status code %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
		return
	}

	// Log the full response body for debugging
	log.Printf("Full response body: %s", body)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
		return
	}

	// Log the structure of the 'articles' field
	articles, ok := result["articles"].([]interface{})
	if !ok {
		log.Fatalf("Error: articles field is not of the expected type, actual type: %T", result["articles"])
		return
	}

	for _, article := range articles {
		articleMap, ok := article.(map[string]interface{})
		if !ok {
			log.Printf("Error: article is not of the expected type")
			continue
		}

		// Safely convert fields to string, handling potential nil values
		title, _ := articleMap["title"].(string)
		description, _ := articleMap["description"].(string)
		category := "business" // Assuming category is constant in your example
		sourceName, _ := articleMap["source"].(map[string]interface{})["name"].(string)
		url, _ := articleMap["url"].(string)

		news := models.News{
			Title:    title,
			Content:  description,
			Category: category,
			Source:   sourceName,
			URL:      url,
		}

		// Use OpenAI API to summarize news articles
		summary, err := SummarizeArticleOpenAI(news.Content)
		if err != nil {
			log.Printf("Error summarizing article: %v", err)
			continue
		}
		news.Content = summary

		if err := db.Create(&news).Error; err != nil {
			log.Printf("Error storing news article: %v", err)
		}
	}
}

func SummarizeArticleOpenAI(articleContent string) (string, error) {
	_, openaiAPIKey := getAPIKeys()

	// Update the model to a supported one, such as gpt-3.5-turbo
	url := "https://api.openai.com/v1/completions"
	data := OpenAIRequest{
		Model:       "gpt-3.5-turbo", // Use a supported model
		Prompt:      fmt.Sprintf("Summarize the following article: %s", articleContent),
		MaxTokens:   150,
		Temperature: 0.7,
	}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response OpenAIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// Log the full OpenAI response for debugging
	log.Printf("OpenAI Response: %s", body)

	// Check if the Choices array is empty
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no summary returned by OpenAI")
	}

	return response.Choices[0].Text, nil
}

// GeneratePersonalizedNewsOpenAI fetches and generates personalized news summaries based on user preferences and interactions.
func GeneratePersonalizedNewsOpenAI(db *gorm.DB, userID string, categories []string, interactions []models.UserInteraction) []models.News {
	var news []models.News

	// Fetch news based on preferred categories
	db.Where("category IN (?)", categories).Find(&news)

	// Example: Apply basic interaction logic (e.g., promote articles based on interactions)
	// This section can be enhanced based on specific personalization needs

	return news
}
