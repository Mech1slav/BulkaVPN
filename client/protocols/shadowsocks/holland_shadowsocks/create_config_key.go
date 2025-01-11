package holland_shadowsocks

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func CreateHollandVPNKey() (string, error) {
	apiURL := Ulr
	method := "POST"

	data := map[string]interface{}{}
	postData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
	reqBody := bytes.NewBuffer(postData)

	req, err := http.NewRequest(method, apiURL, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.SetBasicAuth(User, Password)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // Учитываем и 200, и 201 статус
		return "", fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	accessKey, ok := result["accessUrl"].(string)
	if !ok {
		return "", fmt.Errorf("accessUrl not found in response")
	}

	return accessKey, nil
}
