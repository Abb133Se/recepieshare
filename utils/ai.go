package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CalorieNinjasResponse matches the API response structure
type CalorieNinjasResponse struct {
	Items []struct {
		Name        string  `json:"name"`
		Calories    float64 `json:"calories"`
		ServingSize float64 `json:"serving_size_g"`
	} `json:"items"`
}

func EstimateCalories(apiKey string, ingredients []string) (float64, error) {
	// Join ingredients with comma separator for the query param
	query := strings.Join(ingredients, ", ")
	encodedQuery := url.QueryEscape(query)

	url := "https://api.calorieninjas.com/v1/nutrition?query=" + encodedQuery

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error: %s", string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var cnResp CalorieNinjasResponse
	err = json.Unmarshal(bodyBytes, &cnResp)
	if err != nil {
		return 0, fmt.Errorf("failed to parse API response: %w", err)
	}

	var totalCalories float64
	for _, item := range cnResp.Items {
		totalCalories += item.Calories
	}

	return totalCalories, nil
}
