package edge_tts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type VoiceItem struct {
	Name        string `json:"Name"`
	Gender      string `json:"Gender"`
	Locale      string `json:"Locale"`
	DisplayName string `json:"DisplayName"`
}

func listVoices() ([]VoiceItem, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", VoiceListEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Authority", "speech.platform.bing.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var voices []VoiceItem
	if err := json.Unmarshal(body, &voices); err != nil {
		return nil, err
	}

	return voices, nil
}
