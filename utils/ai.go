package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type NutritionItem struct {
	Name                string  `json:"name"`
	Calories            float64 `json:"calories"`
	ServingSizeG        float64 `json:"serving_size_g"`
	FatTotalG           float64 `json:"fat_total_g"`
	FatSaturatedG       float64 `json:"fat_saturated_g"`
	ProteinG            float64 `json:"protein_g"`
	SodiumMg            float64 `json:"sodium_mg"`
	PotassiumMg         float64 `json:"potassium_mg"`
	CholesterolMg       float64 `json:"cholesterol_mg"`
	CarbohydratesTotalG float64 `json:"carbohydrates_total_g"`
	FiberG              float64 `json:"fiber_g"`
	SugarG              float64 `json:"sugar_g"`
}

func EstimateNutrition(apiKey string, ingredients []string) ([]NutritionItem, error) {
	query := strings.Join(ingredients, ", ")
	url := "https://api.calorieninjas.com/v1/nutrition?query=" + url.QueryEscape(query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Items []NutritionItem `json:"items"`
	}

	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return result.Items, nil
}
