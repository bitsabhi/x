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

	"gorm.io/gorm"
	"newsapp/models"
)

type HuggingFaceRequest struct {
	Inputs string `json:"inputs"`
}

type HuggingFaceResponse struct {
	SummaryText string `json:"summary_text"`
}

// Updated GenerateSummaryHuggingFace to handle array responses
func GenerateSummaryHuggingFace(articleContent string) (string, error) {
	huggingFaceAPIKey := os.Getenv("HUGGINGFACE_API_KEY")
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

	// Handle array response
	var result []HuggingFaceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Return the first summary text in the array
	if len(result) > 0 {
		return result[0].SummaryText, nil
	}

	return "", fmt.Errorf("no summary text found in the response")
}

// FetchAndStoreNewsHuggingFace with The Guardian API
func FetchAndStoreNewsHuggingFace(db *gorm.DB) {
	guardianAPIKey := os.Getenv("GUARDIAN_API_KEY")
	if guardianAPIKey == "" {
		log.Println("The Guardian API key is not set")
		return
	}

	url := fmt.Sprintf("https://content.guardianapis.com/search?section=business&show-fields=bodyText,headline&api-key=%s", guardianAPIKey)
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

	response, ok := result["response"].(map[string]interface{})
	if !ok {
		log.Fatalf("Error: response field is not of the expected type, actual type: %T", result["response"])
		return
	}

	results, ok := response["results"].([]interface{})
	if !ok {
		log.Fatalf("Error: results field is not of the expected type, actual type: %T", response["results"])
		return
	}

	for _, article := range results {
		articleMap, ok := article.(map[string]interface{})
		if !ok {
			log.Printf("Error: article is not of the expected type")
			continue
		}

		fields, fieldsOk := articleMap["fields"].(map[string]interface{})
		if !fieldsOk {
			log.Printf("Error: fields are missing or not of the expected type, skipping article")
			continue
		}

		title, titleOk := fields["headline"].(string)
		articleContent, contentOk := fields["bodyText"].(string)
		url, urlOk := articleMap["webUrl"].(string)
		sourceName := "The Guardian"

		// Skip the article if the most critical fields are missing
		if !titleOk || !urlOk || !contentOk {
			log.Printf("Error: essential article fields missing, skipping article")
			continue
		}

		news := models.News{
			Title:    title,
			Content:  articleContent,
			Category: "business",
			Source:   sourceName,
			URL:      url,
		}

		// Generate summary using Hugging Face
		summary, err := GenerateSummaryHuggingFace(news.Content)
		if err != nil {
			log.Printf("Error summarizing article '%s': %v", news.Title, err)
			continue
		}
		news.Content = summary

		// Attempt to insert the news article into the database
		if err := db.Create(&news).Error; err != nil {
			log.Printf("Error storing news article: %v", err)
		} else {
			log.Printf("Successfully stored article: %s", news.Title)
		}
	}
}

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
