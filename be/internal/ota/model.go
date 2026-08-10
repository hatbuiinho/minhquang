package ota

type Update struct {
	Available        bool   `json:"available"`
	Platform         string `json:"platform,omitempty"`
	Channel          string `json:"channel,omitempty"`
	Version          string `json:"version,omitempty"`
	URL              string `json:"url,omitempty"`
	Checksum         string `json:"checksum,omitempty"`
	Mandatory        bool   `json:"mandatory,omitempty"`
	MinNativeVersion string `json:"min_native_version,omitempty"`
	MaxNativeVersion string `json:"max_native_version,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

type CheckInput struct {
	Platform       string
	Channel        string
	CurrentVersion string
	NativeVersion  string
}
