package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"newsapp/models"
)

var huggingFaceAPIKey = os.Getenv("HUGGINGFACE_API_KEY")

type HuggingFaceRequest struct {
	Inputs string `json:"inputs"`
}

type HuggingFaceResponse struct {
	SummaryText string `json:"summary_text"`
}

// GenerateSummaryHuggingFace calls the Hugging Face API to generate a summary of the given article content
func GenerateSummaryHuggingFace(articleContent string) (string, error) {
	if huggingFaceAPIKey == "" {
		log.Println("Hugging Face API key is not set")
		return "", fmt.Errorf("Hugging Face API key is not set")
	}

	if articleContent == "" {
		log.Println("Article content is empty, skipping summarization")
		return "", fmt.Errorf("article content is empty")
	}

	url := "https://api-inference.huggingface.co/models/facebook/bart-large-cnn"
	requestBody := HuggingFaceRequest{
		Inputs: articleContent,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+huggingFaceAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("Hugging Face API response: %s", body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		SummaryText string `json:"summary_text"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if result.SummaryText == "" {
		return "", fmt.Errorf("no summary returned by Hugging Face")
	}

	return result.SummaryText, nil
}

// GeneratePersonalizedNewsHuggingFace generates personalized news based on user preferences and interactions
func GeneratePersonalizedNewsHuggingFace(db *gorm.DB, userID string, categories []string, interactions []models.UserInteraction) []models.News {
	var news []models.News

	db.Where("category IN (?)", categories).Find(&news)

	for i := range news {
		summary, err := GenerateSummaryHuggingFace(news[i].Content)
		if err != nil {
			log.Printf("Error summarizing article '%s': %v", news[i].Title, err)
			continue
		}
		news[i].Content = summary
	}

	return news
}

// FetchAndStoreNewsHuggingFace fetches news articles, summarizes them using Hugging Face, and stores them in the database
func FetchAndStoreNewsHuggingFace(db *gorm.DB) {
	if db == nil {
		log.Fatalf("Database connection is nil")
		return
	}

	newsAPIKey := os.Getenv("NEWSAPI_KEY")
	if newsAPIKey == "" {
		log.Fatalf("News API key is not set")
		return
	}

	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=us&category=business&apiKey=%s", newsAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching news: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: received status code %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
		return
	}

	log.Printf("Full response body: %s", body)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
		return
	}

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

		title, _ := articleMap["title"].(string)
		description, _ := articleMap["description"].(string)
		source, _ := articleMap["source"].(map[string]interface{})
		sourceName, _ := source["name"].(string)
		url, _ := articleMap["url"].(string)

		if title == "" || description == "" || sourceName == "" || url == "" {
			log.Printf("Error: article fields missing, skipping article")
			continue
		}

		news := models.News{
			Title:    title,
			Content:  description,
			Category: "business",
			Source:   sourceName,
			URL:      url,
		}

		summary, err := GenerateSummaryHuggingFace(news.Content)
		if err != nil {
			log.Printf("Error summarizing article '%s': %v", news.Title, err)
			continue
		}
		news.Content = summary

		if err := db.Create(&news).Error; err != nil {
			log.Printf("Error storing news article: %v", err)
		}
	}
}
