package core

import (
	"fmt"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	ServerName    = "mcp-link"
	ServerVersion = "1.0.0"
)

func LoadOpenAPIDoc(src string) (*libopenapi.DocumentModel[v3high.Document], error) {
	var data []byte
	var err error
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		resp, err := http.Get(src)
		if err != nil {
			return nil, fmt.Errorf("fetch openapi url: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("http error: %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read openapi body: %w", err)
		}
	} else {
		data, err = os.ReadFile(src)
		if err != nil {
			return nil, err
		}
	}

	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, err
	}

	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return model, nil
}

func LoadEnv(key, def string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return val
}
