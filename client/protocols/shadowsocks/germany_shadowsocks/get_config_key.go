package germany_shadowsocks

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetKey(ovpnConfig string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("GET", Ulr, nil)
	if err != nil {
		fmt.Printf("GetKey: failed to create GET request: %v\n", err)
		return false
	}
	req.SetBasicAuth(User, Password)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("GetKey: failed to execute GET request: %v\n", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("GetKey: failed to read response body: %v\n", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GetKey: unexpected response status: %d, body: %s\n", resp.StatusCode, body)
		return false
	}

	var result AccessKeysResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("GetKey: failed to unmarshal response: %v\n", err)
		return false
	}

	for _, key := range result.AccessKeys {
		if key.AccessURL == ovpnConfig {
			return false // Ключ найден, значит он не удален
		}
	}

	return true
}
