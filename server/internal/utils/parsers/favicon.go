package parsers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/OxytocinGroup/theca-v3/internal/repository"
	"golang.org/x/net/html"
)

type IconCandidate struct {
	URL      string
	Priority int
}

// normalizeURL normalizes the URL to be used as a cache key
func normalizeURL(resourceURL string) string {
	if !strings.HasPrefix(resourceURL, "http://") && !strings.HasPrefix(resourceURL, "https://") {
		resourceURL = "https://" + resourceURL
	}

	u, err := url.Parse(resourceURL)
	if err != nil {
		return resourceURL
	}

	return u.Scheme + "://" + u.Host
}

// createHTTPClient creates HTTP client with reasonable defaults
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow up to 10 redirects, но игнорируем редиректы на авторизацию
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}

			// Если редиректит на авторизацию - останавливаемся
			reqURL := req.URL.String()
			if strings.Contains(reqURL, "login") ||
				strings.Contains(reqURL, "signin") ||
				strings.Contains(reqURL, "auth") ||
				strings.Contains(reqURL, "accounts.google.com") {
				return http.ErrUseLastResponse
			}

			return nil
		},
	}
}

// FetchFaviconBase64 extracts favicon for the specified resource and returns it as base64 encoded string.
// If favicon exists in cache, returns it, otherwise downloads and caches it
func FetchFaviconBase64(ctx context.Context, cacheRepo repository.FaviconCacheRepository, resourceURL string) (string, error) {
	normalizedURL := normalizeURL(resourceURL)

	if cacheRepo != nil {
		if cachedFaviconBase64, err := cacheRepo.GetFaviconBase64(ctx, normalizedURL); err == nil && cachedFaviconBase64 != "" {
			return cachedFaviconBase64, nil
		}
	}

	if !strings.HasPrefix(resourceURL, "http://") && !strings.HasPrefix(resourceURL, "https://") {
		resourceURL = "https://" + resourceURL
	}

	// Специальная обработка для известных сервисов
	if faviconURL := getKnownServiceFavicon(resourceURL); faviconURL != "" {
		faviconBase64, err := downloadAndEncodeToBase64(faviconURL)
		if err == nil && faviconBase64 != "" {
			if cacheRepo != nil {
				_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
			}
			return faviconBase64, nil
		}
	}

	client := createHTTPClient()

	// Создаем запрос с User-Agent для получения полного HTML
	req, err := http.NewRequestWithContext(ctx, "GET", resourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем реалистичный User-Agent для обхода блокировок
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Если редирект на авторизацию - пробуем базовый домен
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if strings.Contains(location, "login") || strings.Contains(location, "signin") || strings.Contains(location, "auth") {
			// Пробуем получить favicon напрямую с базового домена
			baseURL, _ := url.Parse(resourceURL)
			if baseURL != nil {
				return tryFaviconFromBaseDomain(ctx, cacheRepo, normalizedURL, baseURL)
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		// Если не удалось получить основную страницу, пробуем базовый домен
		baseURL, _ := url.Parse(resourceURL)
		if baseURL != nil {
			return tryFaviconFromBaseDomain(ctx, cacheRepo, normalizedURL, baseURL)
		}
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	finalURL := resp.Request.URL.String()
	baseURL, err := url.Parse(finalURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Сначала пробуем стандартные местоположения
	standardIconURL := checkStandardFaviconLocations(baseURL)
	if standardIconURL != "" {
		faviconBase64, err := downloadAndEncodeToBase64(standardIconURL)
		if err == nil && faviconBase64 != "" {
			if cacheRepo != nil {
				_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
			}
			return faviconBase64, nil
		}
	}

	// Парсим HTML и ищем иконки в мета-тегах
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	candidates := findIconCandidates(doc, baseURL)
	for _, candidate := range candidates {
		faviconBase64, err := downloadAndEncodeToBase64(candidate.URL)
		if err == nil && faviconBase64 != "" {
			if cacheRepo != nil {
				_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
			}
			return faviconBase64, nil
		}
	}

	// Пробуем регулярные выражения для поиска в HTML
	iconURL := findIconWithRegex(string(body), baseURL)
	if iconURL != "" {
		faviconBase64, err := downloadAndEncodeToBase64(iconURL)
		if err == nil && faviconBase64 != "" {
			if cacheRepo != nil {
				_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
			}
			return faviconBase64, nil
		}
	}

	// Последняя попытка - дефолтная иконка
	defaultIconURL := baseURL.Scheme + "://" + baseURL.Host + "/favicon.ico"
	faviconBase64, err := downloadAndEncodeToBase64(defaultIconURL)
	if err == nil && faviconBase64 != "" {
		if cacheRepo != nil {
			_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
		}
		return faviconBase64, nil
	}

	return "", fmt.Errorf("failed to find or download any valid favicon")
}

// downloadAndEncodeToBase64 downloads an image from URL and converts it to base64
func downloadAndEncodeToBase64(imageURL string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	if len(imageData) == 0 {
		return "", fmt.Errorf("empty image data")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/x-icon"
	}

	base64Data := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Data), nil
}

func checkStandardFaviconLocations(baseURL *url.URL) string {
	standardPaths := []string{
		"/favicon.ico",
		"/apple-touch-icon.png",
		"/apple-touch-icon-120x120.png",
		"/apple-touch-icon-152x152.png",
		"/apple-touch-icon-180x180.png",
		"/apple-touch-icon-precomposed.png",
		"/apple-icon.png",
		"/android-chrome-192x192.png",
		"/icon-192x192.png",
		"/icon.png",
		"/favicon.png",
		"/favicon-32x32.png",
		"/favicon-16x16.png",
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, path := range standardPaths {
		iconURL := baseURL.Scheme + "://" + baseURL.Host + path
		resp, err := client.Head(iconURL)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return iconURL
		}
	}

	return ""
}

// findIconCandidates searches for icon links in HTML document and returns them sorted by priority
func findIconCandidates(doc *html.Node, baseURL *url.URL) []IconCandidate {
	var candidates []IconCandidate

	// Исправленные приоритеты: меньшее число = больший приоритет
	relPriorities := map[string]int{
		"icon":                         1, // Самый высокий приоритет
		"shortcut icon":                2,
		"apple-touch-icon":             3,
		"apple-touch-icon-precomposed": 4,
		"fluid-icon":                   5,
		"mask-icon":                    6,
		"alternate icon":               7, // Самый низкий приоритет
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href, sizes string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "rel":
					rel = strings.ToLower(strings.TrimSpace(attr.Val))
				case "href":
					href = strings.TrimSpace(attr.Val)
				case "sizes":
					sizes = strings.TrimSpace(attr.Val)
				}
			}

			priority, isIcon := relPriorities[rel]
			if isIcon && href != "" {
				iconURL, err := url.Parse(href)
				if err == nil {
					absoluteURL := baseURL.ResolveReference(iconURL).String()

					// Повышаем приоритет для больших размеров
					if sizes != "" && (strings.Contains(sizes, "32x32") || strings.Contains(sizes, "64x64") || strings.Contains(sizes, "128x128")) {
						priority -= 1 // Увеличиваем приоритет
					}

					candidates = append(candidates, IconCandidate{
						URL:      absoluteURL,
						Priority: priority,
					})
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Сортируем по приоритету (меньшее число = больший приоритет)
	for i := range len(candidates) - 1 {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].Priority > candidates[j].Priority {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	return candidates
}

// findIconWithRegex attempts to find icon URLs using regex patterns when HTML parsing fails
func findIconWithRegex(html string, baseURL *url.URL) string {
	patterns := []string{
		// Различные варианты rel="icon"
		`<link[^>]*rel=["'](?:shortcut\s+)?icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["'](?:shortcut\s+)?icon["']`,
		// Apple touch icon
		`<link[^>]*rel=["']apple-touch-icon[^"']*["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["']apple-touch-icon[^"']*["']`,
		// Open Graph image (фолбэк)
		`<meta[^>]*property=["']og:image["'][^>]*content=["']([^"']+)["']`,
		`<meta[^>]*content=["']([^"']+)["'][^>]*property=["']og:image["']`,
		// Twitter card image (фолбэк)
		`<meta[^>]*name=["']twitter:image["'][^>]*content=["']([^"']+)["']`,
		`<meta[^>]*content=["']([^"']+)["'][^>]*name=["']twitter:image["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern) // Case insensitive
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			iconURL, err := url.Parse(matches[1])
			if err != nil {
				continue
			}

			if !iconURL.IsAbs() {
				iconURL = baseURL.ResolveReference(iconURL)
			}

			return iconURL.String()
		}
	}

	return ""
}

// getKnownServiceFavicon returns direct favicon URL for known services
func getKnownServiceFavicon(resourceURL string) string {
	// Парсим URL для определения домена
	u, err := url.Parse(resourceURL)
	if err != nil {
		return ""
	}

	domain := strings.ToLower(u.Host)
	if strings.HasPrefix(domain, "www.") {
		domain = domain[4:]
	}

	// Известные сервисы с прямыми ссылками на фавиконки
	knownServices := map[string]string{
		"gmail.com":         "https://ssl.gstatic.com/ui/v1/icons/mail/rfr/gmail.ico",
		"mail.google.com":   "https://ssl.gstatic.com/ui/v1/icons/mail/rfr/gmail.ico",
		"google.com":        "https://www.google.com/favicon.ico",
		"youtube.com":       "https://www.youtube.com/favicon.ico",
		"github.com":        "https://github.com/favicon.ico",
		"stackoverflow.com": "https://cdn.sstatic.net/Sites/stackoverflow/Img/favicon.ico",
		"twitter.com":       "https://abs.twimg.com/favicons/twitter.ico",
		"facebook.com":      "https://static.xx.fbcdn.net/rsrc.php/yV/r/hzMapiNYYpW.ico",
		"linkedin.com":      "https://static.licdn.com/sc/h/1bt1uwq5akv756knzdj4l6cdc",
	}

	if faviconURL, exists := knownServices[domain]; exists {
		return faviconURL
	}

	return ""
}

// tryFaviconFromBaseDomain tries to get favicon directly from base domain without redirects
func tryFaviconFromBaseDomain(ctx context.Context, cacheRepo repository.FaviconCacheRepository, normalizedURL string, baseURL *url.URL) (string, error) {
	// Сначала проверяем известные сервисы
	if faviconURL := getKnownServiceFavicon(baseURL.String()); faviconURL != "" {
		faviconBase64, err := downloadAndEncodeToBase64(faviconURL)
		if err == nil && faviconBase64 != "" {
			if cacheRepo != nil {
				_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
			}
			return faviconBase64, nil
		}
	}

	// Пробуем стандартные местоположения
	standardIconURL := checkStandardFaviconLocations(baseURL)
	if standardIconURL != "" {
		faviconBase64, err := downloadAndEncodeToBase64(standardIconURL)
		if err == nil && faviconBase64 != "" {
			if cacheRepo != nil {
				_ = cacheRepo.StoreFaviconBase64(ctx, normalizedURL, faviconBase64)
			}
			return faviconBase64, nil
		}
	}

	return "", fmt.Errorf("failed to find favicon from base domain")
}

func FetchFavicon(ctx context.Context, cacheRepo repository.FaviconCacheRepository, resourceURL string) (string, error) {
	_, err := FetchFaviconBase64(ctx, cacheRepo, resourceURL)
	if err != nil {
		return "", err
	}

	return "", nil
}
