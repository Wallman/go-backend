package mistral

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type Mistral struct {
	baseURL string
	http    *http.Client
}

type TranscribeResponse struct {
	Text  string `json:"text"`
	Usage Usage  `json:"usage"`
}

type Usage struct {
	TotalTokens int `json:"total_tokens"`
}

func (m Mistral) Transcribe(ctx context.Context, uri string) (TranscribeResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("model", "voxtral-mini-latest"); err != nil {
		writer.Close()
		return TranscribeResponse{}, err
	}
	if err := writer.WriteField("file_url", uri); err != nil {
		writer.Close()
		return TranscribeResponse{}, err
	}
	if err := writer.Close(); err != nil {
		return TranscribeResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return TranscribeResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+os.Getenv("MISTRAL_API_KEY"))
	resp, err := m.http.Do(req)
	if err != nil {
		return TranscribeResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return TranscribeResponse{}, fmt.Errorf("non-OK HTTP response: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var response TranscribeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return TranscribeResponse{}, err
	}
	return response, nil
}

func NewMistral(baseURL string) *Mistral {
	return &Mistral{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}
