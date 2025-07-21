package replacepathregex

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	ReplacedPathHeader = "X-Replaced-Path"
)

type Config struct {
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	regex, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, err
	}

	return &ReplacePathRegexHandler{
		next:        next,
		replacement: strings.TrimSpace(config.Replacement),
		regex:       regex,
	}, nil
}

type ReplacePathRegexHandler struct {
	next        http.Handler
	replacement string
	regex       *regexp.Regexp
}

func (h *ReplacePathRegexHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	backReferencesRegex := regexp.MustCompile(`\$\d+`)

	slog.Debug("Request Path: " + req.URL.Path)

	if h.regex != nil && len(h.replacement) > 0 {
		req.Header.Add(ReplacedPathHeader, req.URL.Path)

		if matches := h.regex.FindStringSubmatch(req.URL.String()); len(matches) > 0 {
			matches = matches[1:]

			replacement := h.replacement

			for i, match := range matches {
				replacement = strings.ReplaceAll(replacement, "$"+strconv.Itoa(i+1), match)
			}

			req.URL.Path = backReferencesRegex.ReplaceAllString(replacement, "")
		}
	}

	h.next.ServeHTTP(w, req)
}
