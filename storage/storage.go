// Package storage defines the Storage interface for file upload/download
// operations, along with the request and response DTOs.
//
// The interface is backend-agnostic. Use the s3 sub-package for an
// implementation that proxies uploads/downloads through an S3-compatible
// external service.
package storage

import (
	"context"
	"io"
)

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// Storage is the abstraction for file storage operations.
type Storage interface {
	// Upload stores a file and returns the resulting file name / path.
	Upload(ctx context.Context, req *UploadRequest) (*UploadResponse, error)

	// Download retrieves a file by name and returns a readable stream.
	// The caller is responsible for closing DownloadResponse.Data.
	Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error)
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// UploadRequest contains the data needed to upload a file.
type UploadRequest struct {
	// Folder is the optional sub-folder / prefix within the bucket.
	Folder string

	// Base64Data is the file content encoded as a base64 string.
	Base64Data string
}

// UploadResponse contains the result of a successful upload.
type UploadResponse struct {
	// FileName is the stored file name or full path within the bucket.
	FileName string
}

// DownloadRequest identifies the file to download.
type DownloadRequest struct {
	// FileName is the full file path/name as returned by Upload.
	FileName string
}

// DownloadResponse wraps the downloaded file data.
type DownloadResponse struct {
	// FileName echoes back the requested file name.
	FileName string

	// Data is the file content stream. The caller must close it when done.
	Data io.ReadCloser
}
