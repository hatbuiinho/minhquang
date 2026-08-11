package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const presignExpiry = 5 * time.Minute

type MinIO struct {
	client     *minio.Client
	bucket     string
	endpoint   string
	useSSL     bool
	publicBase string
}

type Config struct {
	Endpoint, AccessKey, SecretKey, Bucket, Region, PublicBase string
	UseSSL                                                     bool
}

type PresignedUpload struct {
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"object_key"`
	UploadURL string    `json:"upload_url"`
	PublicURL string    `json:"public_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewMinIO(cfg Config) (*MinIO, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("minio endpoint, bucket and credentials are required")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &MinIO{client: client, bucket: cfg.Bucket, endpoint: cfg.Endpoint, useSSL: cfg.UseSSL, publicBase: strings.TrimRight(cfg.PublicBase, "/")}, nil
}

func (m *MinIO) PresignAvatar(ctx context.Context, userID, fileName string) (PresignedUpload, error) {
	key := path.Join("avatars", userID, randomID()+"-"+sanitizeFileName(fileName))
	u, err := m.client.PresignedPutObject(ctx, m.bucket, key, presignExpiry)
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("presign avatar upload: %w", err)
	}
	return PresignedUpload{Bucket: m.bucket, ObjectKey: key, UploadURL: u.String(), PublicURL: m.PublicURL(key), ExpiresAt: time.Now().Add(presignExpiry)}, nil
}

func (m *MinIO) PublicURL(key string) string {
	encoded := encodePath(key)
	if m.publicBase != "" {
		return m.publicBase + "/" + encoded
	}
	scheme := "http"
	if m.useSSL {
		scheme = "https"
	}
	return scheme + "://" + m.endpoint + "/" + m.bucket + "/" + encoded
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(path.Base(name), " ", "_"))
	if name == "" || name == "." {
		return "avatar.jpg"
	}
	return url.PathEscape(name)
}

func encodePath(value string) string {
	parts := strings.Split(value, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func randomID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return hex.EncodeToString(value[:])
}
