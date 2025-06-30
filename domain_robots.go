package domainrobots

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Config struct {
	S3Bucket     string `json:"s3Bucket,omitempty"`
	S3Region     string `json:"s3Region,omitempty"`
	S3Endpoint   string `json:"s3Endpoint,omitempty"`
	S3PrefixPath string `json:"s3PrefixPath,omitempty"`

	RobotsTxtPath string `json:"robotsTxtPath,omitempty"`
	SitemapPath   string `json:"sitemapPath,omitempty"`

	Protocol string `json:"protocol,omitempty"`
}

// https://<bucket>.s3.<region>.amazonaws.com/robots_txt/<domain>/robots.txt
// https://<bucket>.s3.<region>.amazonaws.com/sitemap/<domain>/*

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Protocol == "" {
		config.Protocol = "https"
	}

	if config.S3Bucket == "" {
		return nil, fmt.Errorf("s3Bucket is required")
	}

	if config.RobotsTxtPath == "" && config.SitemapPath == "" {
		return nil, fmt.Errorf("robotsTxtPath or sitemapPath is required")
	}

	return &DomainRobotsHandler{
		next:   next,
		config: config,
	}, nil
}

type DomainRobotsHandler struct {
	next   http.Handler
	config *Config
}

func (h *DomainRobotsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originalHost := r.Host

	domain := strings.Split(originalHost, ":")[0]

	targetURL := h.getTargetS3URL(domain)

	if targetURL == "" {
		h.next.ServeHTTP(w, r)
		return
	}

	r.Host = targetURL

	h.next.ServeHTTP(w, r)
}

func (h *DomainRobotsHandler) getTargetS3URL(domain string) string {

	textFilePath := ""

	if h.config.RobotsTxtPath != "" {
		textFilePath = h.config.RobotsTxtPath
	} else {
		textFilePath = h.config.SitemapPath
	}

	if h.config.S3PrefixPath == "" {
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/%s", h.config.S3Bucket, h.config.S3Region, domain, textFilePath)
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/%s/%s", h.config.S3Bucket, h.config.S3Region, h.config.S3PrefixPath, domain, textFilePath)
}
