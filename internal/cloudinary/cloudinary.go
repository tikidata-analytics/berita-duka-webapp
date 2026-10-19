package cloudinary

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tikidata-analytics/berita-duka-webapp/internal/config"
)

// Client handles interaction with Cloudinary Image Upload API.
type Client struct {
	cloudName    string
	apiKey       string
	apiSecret    string
	uploadPreset string
	folder       string
	httpClient   *http.Client
}

// UploadResponse represents the response payload from Cloudinary API.
type UploadResponse struct {
	SecureURL    string `json:"secure_url"`
	PublicID     string `json:"public_id"`
	Format       string `json:"format"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Bytes        int64  `json:"bytes"`
	ErrorMessage string `json:"error,omitempty"`
}

type cloudinaryErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Cloudinary client from configuration.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		cloudName:    cfg.CloudinaryCloudName,
		apiKey:       cfg.CloudinaryAPIKey,
		apiSecret:    cfg.CloudinaryAPISecret,
		uploadPreset: cfg.CloudinaryUploadPreset,
		folder:       cfg.CloudinaryFolder,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if either API Key/Secret or Upload Preset is provided.
func (c *Client) IsConfigured() bool {
	if c.cloudName == "" {
		return false
	}
	return (c.apiKey != "" && c.apiSecret != "") || c.uploadPreset != ""
}

// UploadImage uploads an image file stream to Cloudinary and returns the HTTPS secure URL.
func (c *Client) UploadImage(ctx context.Context, fileReader io.Reader, filename string) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("Cloudinary belum dikonfigurasi. Harap tentukan CLOUDINARY_CLOUD_NAME dan CLOUDINARY_API_KEY/SECRET di .env")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Mode 1: Signed Upload (Preferred for secure server-side execution)
	if c.apiKey != "" && c.apiSecret != "" {
		params := map[string]string{
			"timestamp": timestamp,
		}
		if c.folder != "" {
			params["folder"] = c.folder
		}

		signature := generateSignature(params, c.apiSecret)

		_ = writer.WriteField("api_key", c.apiKey)
		_ = writer.WriteField("timestamp", timestamp)
		_ = writer.WriteField("signature", signature)
		if c.folder != "" {
			_ = writer.WriteField("folder", c.folder)
		}
	} else if c.uploadPreset != "" {
		// Mode 2: Unsigned Upload with Preset
		_ = writer.WriteField("upload_preset", c.uploadPreset)
		if c.folder != "" {
			_ = writer.WriteField("folder", c.folder)
		}
	}

	// Attach file part
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, fileReader); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", c.cloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Cloudinary API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from Cloudinary: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp cloudinaryErrorResponse
		if err := json.Unmarshal(respBytes, &errResp); err == nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("Cloudinary error: %s", errResp.Error.Message)
		}
		return "", fmt.Errorf("Cloudinary returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var uploadResp UploadResponse
	if err := json.Unmarshal(respBytes, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse Cloudinary response: %w", err)
	}

	if uploadResp.SecureURL == "" {
		return "", fmt.Errorf("Cloudinary response did not contain secure_url")
	}

	return uploadResp.SecureURL, nil
}

// generateSignature computes the SHA-1 signature according to Cloudinary specifications.
// Parameters are sorted alphabetically and joined with '&', followed by api_secret.
func generateSignature(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params[k]))
	}

	toSign := strings.Join(pairs, "&") + secret
	hash := sha1.Sum([]byte(toSign))
	return fmt.Sprintf("%x", hash)
}
