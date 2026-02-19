package nexus

// Keep models intentionally minimal.
// Add fields only as needed to avoid tight coupling to API shape.

type ValidateUserResponse struct {
	UserID int    `json:"user_id"`
	Key    string `json:"key,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`

	IsPremium   bool `json:"is_premium,omitempty"`
	IsSupporter bool `json:"is_supporter,omitempty"`
}

type ModInfo struct {
	ModID int    `json:"mod_id"`
	Name  string `json:"name,omitempty"`

	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`

	Version string `json:"version,omitempty"`
	Author  string `json:"author,omitempty"`

	CreatedTime string `json:"created_time,omitempty"`
	UpdatedTime string `json:"updated_time,omitempty"`

	EndorsementCount int `json:"endorsement_count,omitempty"`
	UniqueDownloads  int `json:"unique_downloads,omitempty"`
	Downloads        int `json:"downloads,omitempty"`
}

// NOTE: The "files" endpoint returns { "files": [ ... ] } and the file payload uses
// fields like: file_id, file_name, uploaded_timestamp, size, etc.
type FileInfo struct {
	FileID int    `json:"file_id"`
	Name   string `json:"name,omitempty"`

	FileName string `json:"file_name,omitempty"`

	Version string `json:"version,omitempty"`

	CategoryName string `json:"category_name,omitempty"`
	CategoryID   int    `json:"category_id,omitempty"`

	// API returns "size" (bytes).
	Size int64 `json:"size,omitempty"`

	UploadedTimestamp int64  `json:"uploaded_timestamp,omitempty"`
	UploadedTime      string `json:"uploaded_time,omitempty"`
	UpdatedTime       string `json:"updated_time,omitempty"`

	Description string `json:"description,omitempty"`

	IsPrimary bool `json:"is_primary,omitempty"`
}

// DownloadLink is returned by download_link.json endpoints.
type DownloadLink struct {
	Name      string `json:"name,omitempty"`
	ShortName string `json:"short_name,omitempty"`
	URI       string `json:"URI"`
}

// SearchResult is kept for future phases (GraphQL/web search).
type SearchResult struct {
	ModID int    `json:"mod_id"`
	Name  string `json:"name,omitempty"`

	Summary     string `json:"summary,omitempty"`
	Author      string `json:"author,omitempty"`
	Version     string `json:"version,omitempty"`
	UpdatedTime string `json:"updated_time,omitempty"`
}
