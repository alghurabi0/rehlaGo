package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func wistiaReq(method, url string, jsonData []byte) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("couldn't get new req: %v\n", err)
		return "", err
	}
	token := os.Getenv("wistia_token")
	if token == "" {
		fmt.Println("wistia_token environment variable is not set")
		return "", fmt.Errorf("missing wistia_token")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to send a req: %v\n", err)
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		fmt.Printf("unexpected status code: %d\n", res.StatusCode)
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("response body: %s\n", body)
		return "", fmt.Errorf("received non-2xx response code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("failed to read body: %v\n", err)
		return "", err
	}
	fmt.Printf("response body: %s\n", body)
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("failed to unmarshal data: %v\n", err)
		return "", err
	}
	hashedId := response["hashedId"].(string)
	if hashedId == "" {
		return "", nil
	}
	return hashedId, nil
}
