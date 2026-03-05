// Package s3 provides an S3 proxy implementation of storage.Storage.
//
// It communicates with an external S3 file service via HTTP/JSON (upload at
// /simpanfile, download at /ambilfile). This is not a direct AWS S3 SDK
// implementation but rather a proxy adapter for services that wrap S3.
//
// Usage:
//
//	cfg := s3.Config{
//	    BaseURL:         "https://file-service.internal",
//	    AccessKeyID:     "...",
//	    BucketName:      "my-bucket",
//	    Endpoint:        "s3.amazonaws.com",
//	    SecretAccessKey: "...",
//	}
//
//	store := s3.New(cfg)
//	resp, err := store.Upload(ctx, &storage.UploadRequest{
//	    Folder:     "images",
//	    Base64Data: base64EncodedFile,
//	})
package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/semmidev/bee-go-kit/storage"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	// ErrMarshalRequest is returned when the upload/download request body
	// cannot be marshalled to JSON.
	ErrMarshalRequest = fmt.Errorf("s3: failed to marshal request body")

	// ErrCreateRequest is returned when the HTTP request cannot be created.
	ErrCreateRequest = fmt.Errorf("s3: failed to create HTTP request")

	// ErrSendRequest is returned when the HTTP request fails to send.
	ErrSendRequest = fmt.Errorf("s3: failed to send HTTP request")

	// ErrNon200Response is returned when the external service responds with
	// a non-200 status code.
	ErrNon200Response = fmt.Errorf("s3: external service returned non-200 status")

	// ErrDecodeResponse is returned when the service response cannot be
	// decoded as JSON.
	ErrDecodeResponse = fmt.Errorf("s3: failed to decode response")

	// ErrUploadFailed is returned when the service reports a non-200
	// response code in its own payload.
	ErrUploadFailed = fmt.Errorf("s3: upload failed")

	// ErrInvalidBase64 is returned when the provided base64 data is empty
	// or too short to be valid.
	ErrInvalidBase64 = fmt.Errorf("s3: invalid base64 data")

	// ErrEmptyResponse is returned when the service response description
	// is empty.
	ErrEmptyResponse = fmt.Errorf("s3: empty response received")
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds the parameters needed to connect to the S3 proxy service.
type Config struct {
	// BaseURL is the root URL of the file service (e.g. "https://files.internal").
	BaseURL string

	// AccessKeyID is the S3 access key.
	AccessKeyID string

	// BucketName is the target S3 bucket name.
	BucketName string

	// Endpoint is the S3-compatible endpoint URL.
	Endpoint string

	// SecretAccessKey is the S3 secret key.
	SecretAccessKey string

	// DefaultFolder is an optional default folder prefix for uploads.
	DefaultFolder string

	// Timeout is the HTTP client timeout. Defaults to 60s if zero.
	Timeout time.Duration
}

// ---------------------------------------------------------------------------
// S3Storage
// ---------------------------------------------------------------------------

// S3Storage implements storage.Storage by proxying file uploads and downloads
// through an external HTTP service.
type S3Storage struct {
	baseURL         string
	accessKeyID     string
	bucketName      string
	endpoint        string
	secretAccessKey string
	httpClient      *http.Client
}

// New creates a new S3Storage from the given configuration.
func New(cfg Config) storage.Storage {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &S3Storage{
		baseURL:         cfg.BaseURL,
		accessKeyID:     cfg.AccessKeyID,
		bucketName:      cfg.BucketName,
		endpoint:        cfg.Endpoint,
		secretAccessKey: cfg.SecretAccessKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ---------------------------------------------------------------------------
// Internal request/response types
// ---------------------------------------------------------------------------

type uploadRequest struct {
	AccessKeyID     string `json:"accesskeyid"`
	BucketName      string `json:"bucketname"`
	Endpoint        string `json:"endpoint"`
	File            string `json:"file"`
	Folder          string `json:"folder"`
	SecretAccessKey string `json:"secretaccesskey"`
}

type uploadResponse struct {
	ResponseCode        int    `json:"responseCode"`
	ResponseDescription string `json:"responseDescription"`
	Data                string `json:"data"`
}

type downloadRequest struct {
	AccessKeyID     string `json:"accesskeyid"`
	BucketName      string `json:"bucketname"`
	Endpoint        string `json:"endpoint"`
	File            string `json:"file"`
	Folder          string `json:"folder"`
	SecretAccessKey string `json:"secretaccesskey"`
}

// ---------------------------------------------------------------------------
// storage.Storage implementation
// ---------------------------------------------------------------------------

// Upload stores a base64-encoded file via the external service.
func (s *S3Storage) Upload(ctx context.Context, req *storage.UploadRequest) (*storage.UploadResponse, error) {
	// Validate input.
	if err := s.validateInput(req.Base64Data); err != nil {
		return nil, err
	}

	folder := strings.TrimSpace(req.Folder)
	if folder != "" {
		folder += "/" // the service expects a trailing slash
	}

	// Build and send the request.
	body, err := s.prepareUploadBody(req.Base64Data, folder)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMarshalRequest, err)
	}

	httpReq, err := s.newRequest(ctx, "simpanfile", body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateRequest, err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendRequest, err)
	}
	defer resp.Body.Close()

	// Parse and validate the response.
	cleanFileName, err := s.parseUploadResponse(resp)
	if err != nil {
		return nil, err
	}

	// No folder → return file name only.
	if folder == "" {
		return &storage.UploadResponse{FileName: cleanFileName}, nil
	}

	// With folder → build the full path with query string.
	finalPath := fmt.Sprintf("%s/%s?folder=%s", s.bucketName, cleanFileName, strings.TrimSuffix(folder, "/"))
	return &storage.UploadResponse{FileName: finalPath}, nil
}

// Download retrieves a file via the external service. The caller must close
// the returned DownloadResponse.Data when done.
func (s *S3Storage) Download(ctx context.Context, req *storage.DownloadRequest) (*storage.DownloadResponse, error) {
	fileName, folder := s.parseDownloadPath(req.FileName)

	body, err := json.Marshal(downloadRequest{
		AccessKeyID:     s.accessKeyID,
		BucketName:      s.bucketName,
		Endpoint:        s.endpoint,
		File:            fileName,
		Folder:          folder,
		SecretAccessKey: s.secretAccessKey,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMarshalRequest, err)
	}

	httpReq, err := s.newRequest(ctx, "ambilfile", body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateRequest, err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendRequest, err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrNon200Response, resp.StatusCode, string(bodyBytes))
	}

	return &storage.DownloadResponse{
		FileName: req.FileName,
		Data:     resp.Body,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// validateInput ensures the base64 data is neither empty nor suspiciously short.
func (s *S3Storage) validateInput(base64Data string) error {
	if base64Data == "" || len(base64Data) < 10 {
		return ErrInvalidBase64
	}
	return nil
}

// prepareUploadBody marshals the upload request to JSON bytes.
func (s *S3Storage) prepareUploadBody(base64Data, folder string) ([]byte, error) {
	return json.Marshal(uploadRequest{
		AccessKeyID:     s.accessKeyID,
		BucketName:      s.bucketName,
		Endpoint:        s.endpoint,
		File:            base64Data,
		Folder:          folder,
		SecretAccessKey: s.secretAccessKey,
	})
}

// newRequest builds an HTTP POST request for the given service path.
func (s *S3Storage) newRequest(ctx context.Context, path string, body []byte) (*http.Request, error) {
	u := fmt.Sprintf("%s/%s", strings.TrimRight(s.baseURL, "/"), strings.TrimLeft(path, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// parseUploadResponse reads and validates the upload response.
func (s *S3Storage) parseUploadResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%w: status %d, body: %s", ErrNon200Response, resp.StatusCode, string(bodyBytes))
	}

	var res uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}

	if res.ResponseCode != 200 {
		return "", fmt.Errorf("%w: response code %d, description: %s", ErrUploadFailed, res.ResponseCode, res.ResponseDescription)
	}

	if res.ResponseDescription == "" {
		return "", ErrEmptyResponse
	}

	return s.extractFileName(res.ResponseDescription), nil
}

// extractFileName strips common path prefixes from the service response.
func (s *S3Storage) extractFileName(responseDesc string) string {
	name := strings.TrimPrefix(responseDesc, "./file/")
	name = strings.TrimPrefix(name, "/file/")
	name = strings.TrimPrefix(name, "file/")
	return name
}

// parseDownloadPath separates the file name and folder from a stored path.
//
// Input formats:
//
//	"bucket/file.png?folder=icons"  → file="file.png", folder="icons"
//	"bucket/file.png"               → file="file.png", folder=""
func (s *S3Storage) parseDownloadPath(fullPath string) (fileName, folder string) {
	// Remove bucket prefix.
	pathWithoutBucket := fullPath
	if idx := strings.Index(fullPath, "/"); idx != -1 {
		pathWithoutBucket = fullPath[idx+1:]
	}

	// Split on "?" to extract folder from query string.
	if strings.Contains(pathWithoutBucket, "?") {
		parts := strings.SplitN(pathWithoutBucket, "?", 2)
		fileName = parts[0]
		if params, err := url.ParseQuery(parts[1]); err == nil {
			folder = params.Get("folder")
		}
	} else {
		fileName = pathWithoutBucket
	}

	return fileName, folder
}
