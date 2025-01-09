package holland_shadowsocks

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AccessKey struct {
	ID        string `json:"id"`
	AccessURL string `json:"accessUrl"`
}

type AccessKeysResponse struct {
	AccessKeys []AccessKey `json:"accessKeys"`
}

func DeleteKeyByConfig(ovpnConfig string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("GET", Ulr, nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request: %v", err)
	}
	req.SetBasicAuth(User, Password)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute GET request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, body)
	}

	var result AccessKeysResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var keyID string
	for _, key := range result.AccessKeys {
		if key.AccessURL == ovpnConfig {
			keyID = key.ID
			break
		}
	}

	if keyID == "" {
		return fmt.Errorf("client with config %s not found", ovpnConfig)
	}

	deleteURL := fmt.Sprintf("%s/%s", Ulr, keyID)
	delReq, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %v", err)
	}
	delReq.SetBasicAuth(User, Password)

	delResp, err := client.Do(delReq)
	if err != nil {
		return fmt.Errorf("failed to execute DELETE request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(delResp.Body)

	if delResp.StatusCode != http.StatusNoContent {
		delBody, _ := io.ReadAll(delResp.Body)
		return fmt.Errorf("unexpected response status: %d, body: %s", delResp.StatusCode, delBody)
	}

	return nil
}
