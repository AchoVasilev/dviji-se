package middleware

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var assetHashes map[string]string

// InitAssetHashes walks staticDir, computes a SHA1 hash for each file,
// and stores the first 12 hex chars keyed by /static/... URL path.
func InitAssetHashes(staticDir string) {
	hashes := make(map[string]string)

	err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}

		rel, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}

		urlPath := "/static/" + filepath.ToSlash(rel)
		hash := hex.EncodeToString(h.Sum(nil))[:12]
		hashes[urlPath] = hash

		return nil
	})

	if err != nil {
		slog.Error("Failed to compute asset hashes", "error", err)
	}

	assetHashes = hashes
	slog.Info(fmt.Sprintf("Computed hashes for %d static assets", len(hashes)))
}

// AssetURL returns path?v=<hash> if a hash exists, otherwise path unchanged.
func AssetURL(path string) string {
	if h, ok := assetHashes[path]; ok {
		return path + "?v=" + h
	}
	return path
}

