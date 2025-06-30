package domainrobots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config should work",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			wantErr: false,
		},
		{
			name: "config with custom paths should work",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "custom-robots.txt",
			},
			wantErr: false,
		},
		{
			name: "config with S3PrefixPath should work",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				S3PrefixPath:  "assets",
				RobotsTxtPath: "robots.txt",
			},
			wantErr: false,
		},
		{
			name: "missing S3Bucket should fail",
			config: &Config{
				S3Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "s3Bucket is required",
		},
		{
			name: "empty S3Bucket should fail",
			config: &Config{
				S3Bucket: "",
				S3Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "s3Bucket is required",
		},
		{
			name: "missing RobotsTxtPath and SitemapPath should fail",
			config: &Config{
				S3Bucket: "test-bucket",
				S3Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "robotsTxtPath or sitemapPath is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), tt.config, "test")

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}

			if handler == nil {
				t.Error("New() returned nil handler")
			}

		})
	}
}

func TestDomainRobotsHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		requestPath  string
		requestHost  string
		expectedHost string
		nextCalled   bool
	}{
		{
			name: "basic request with robots.txt",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			requestPath:  "/robots.txt",
			requestHost:  "example.com",
			expectedHost: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/robots.txt",
			nextCalled:   true,
		},
		{
			name: "request with sitemap.xml",
			config: &Config{
				S3Bucket:    "test-bucket",
				S3Region:    "us-east-1",
				SitemapPath: "sitemap.xml",
			},
			requestPath:  "/sitemap.xml",
			requestHost:  "example.com",
			expectedHost: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/sitemap.xml",
			nextCalled:   true,
		},
		{
			name: "request with custom robots path",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "custom-robots.txt",
			},
			requestPath:  "/custom-robots.txt",
			requestHost:  "example.com",
			expectedHost: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/custom-robots.txt",
			nextCalled:   true,
		},
		{
			name: "request with S3PrefixPath",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				S3PrefixPath:  "assets",
				RobotsTxtPath: "robots.txt",
			},
			requestPath:  "/robots.txt",
			requestHost:  "example.com",
			expectedHost: "https://test-bucket.s3.us-east-1.amazonaws.com/assets/example.com/robots.txt",
			nextCalled:   true,
		},
		{
			name: "request with port in host",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			requestPath:  "/robots.txt",
			requestHost:  "example.com:8080",
			expectedHost: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/robots.txt",
			nextCalled:   true,
		},
		{
			name: "non-robots request should pass through",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			requestPath: "/other/path",
			requestHost: "example.com",
			nextCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			var capturedHost string

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				capturedHost = r.Host
				w.WriteHeader(http.StatusOK)
			})

			handler, err := New(context.Background(), nextHandler, tt.config, "test")
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			req.Host = tt.requestHost
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if !nextCalled {
				t.Error("Expected next handler to be called")
			}

			if tt.expectedHost != "" && capturedHost != tt.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tt.expectedHost, capturedHost)
			}
		})
	}
}

func TestGetTargetS3URL(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		domain   string
		expected string
	}{
		{
			name: "basic S3 URL without prefix",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			domain:   "example.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/robots.txt",
		},
		{
			name: "S3 URL with prefix path",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				S3PrefixPath:  "assets",
				RobotsTxtPath: "robots.txt",
			},
			domain:   "example.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/assets/example.com/robots.txt",
		},
		{
			name: "S3 URL with sitemap path",
			config: &Config{
				S3Bucket:    "test-bucket",
				S3Region:    "us-east-1",
				SitemapPath: "sitemap.xml",
			},
			domain:   "example.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/sitemap.xml",
		},
		{
			name: "S3 URL with custom region",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "eu-west-1",
				RobotsTxtPath: "robots.txt",
			},
			domain:   "example.com",
			expected: "https://test-bucket.s3.eu-west-1.amazonaws.com/example.com/robots.txt",
		},
		{
			name: "S3 URL with complex domain",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			domain:   "sub.example.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/sub.example.com/robots.txt",
		},
		{
			name: "S3 URL with prefix and sitemap",
			config: &Config{
				S3Bucket:     "test-bucket",
				S3Region:     "us-east-1",
				S3PrefixPath: "static",
				SitemapPath:  "sitemap.xml",
			},
			domain:   "example.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/static/example.com/sitemap.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &DomainRobotsHandler{
				config: tt.config,
			}

			result := handler.getTargetS3URL(tt.domain)
			if result != tt.expected {
				t.Errorf("getTargetS3URL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetTargetS3URL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		domain   string
		expected string
	}{
		{
			name: "empty robots path should use sitemap path",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "",
				SitemapPath:   "sitemap.xml",
			},
			domain:   "example.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/example.com/sitemap.xml",
		},
		{
			name: "empty domain",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			domain:   "",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com//robots.txt",
		},
		{
			name: "domain with special characters",
			config: &Config{
				S3Bucket:      "test-bucket",
				S3Region:      "us-east-1",
				RobotsTxtPath: "robots.txt",
			},
			domain:   "test-domain.com",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/test-domain.com/robots.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &DomainRobotsHandler{
				config: tt.config,
			}

			result := handler.getTargetS3URL(tt.domain)
			if result != tt.expected {
				t.Errorf("getTargetS3URL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestServeHTTP_NextHandlerCalled(t *testing.T) {
	// 測試當請求不是 robots.txt 或 sitemap.xml 時，應該調用下一個處理器
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	config := &Config{
		S3Bucket:      "test-bucket",
		S3Region:      "us-east-1",
		RobotsTxtPath: "robots.txt",
	}

	handler, err := New(context.Background(), nextHandler, config, "test")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/other/path", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("Expected next handler to be called for non-robots/sitemap requests")
	}
}

// 基準測試
func BenchmarkDomainRobotsHandler_ServeHTTP(b *testing.B) {
	config := &Config{
		S3Bucket: "test-bucket",
		S3Region: "us-east-1",
	}

	handler, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "test")
	if err != nil {
		b.Fatalf("Failed to create handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/robots.txt", nil)
	req.Host = "example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkGetTargetS3URL(b *testing.B) {
	handler := &DomainRobotsHandler{
		config: &Config{
			S3Bucket:      "test-bucket",
			S3Region:      "us-east-1",
			RobotsTxtPath: "/robots.txt",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.getTargetS3URL("example.com")
	}
}
