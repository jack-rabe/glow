package docs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	htm "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

func GetPage(url string) (string, error) {
	// Create HTTP client with timeout and redirect support
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create request for %s: %v", url, err)
		log.Println(errorMsg)
		return "", fmt.Errorf(errorMsg)
	}

	req.Header.Set("User-Agent", USER_AGENT)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch %s: %v", url, err)
		log.Println(errorMsg)
		return "", fmt.Errorf(errorMsg)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode >= 400 {
		errorMsg := fmt.Sprintf("Failed to fetch %s - status code %d", url, resp.StatusCode)
		log.Println(errorMsg)
		return "", fmt.Errorf(errorMsg)
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("Error reading response body from %s: %v", url, err)
		log.Println(errorMsg)
		return "", fmt.Errorf(errorMsg)
	}

	fileName := strings.TrimSuffix(url, ".html")
	fileName = strings.TrimPrefix(fileName, "https://docs.aws.amazon.com/")
	fileName = strings.ReplaceAll(fileName, "/", "-")
	fileName += ".md"
	markdown := extractContentFromHTML(string(bodyBytes))
	err = os.WriteFile(fileName, []byte(markdown), 0777)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func extractContentFromHTML(html string) string {
	if strings.TrimSpace(html) == "" {
		return "<e>Empty HTML content</e>" // todo - fix
	}

	// Parse HTML using goquery (like BeautifulSoup)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return fmt.Sprintf("<e>Error parsing HTML: %v</e>", err)
	}

	var mainContent *goquery.Selection

	// Common content container selectors
	contentSelectors := []string{
		"main",
		"article",
		"#main-content",
		".main-content",
		"#content",
		".content",
		"div[role='main']",
		"#awsdocs-content",
		".awsui-article",
	}

	for _, selector := range contentSelectors {
		selection := doc.Find(selector).First()
		if selection.Length() > 0 {
			mainContent = selection
			break
		}
	}

	// Fallback to <body> or entire document
	if mainContent == nil {
		mainContent = doc.Find("body")
		if mainContent.Length() == 0 {
			mainContent = doc.Selection
		}
	}

	// Remove unwanted navigation/utility elements
	navSelectors := []string{
		"noscript",
		".prev-next",
		"#main-col-footer",
		".awsdocs-page-utilities",
		"#quick-feedback-yes",
		"#quick-feedback-no",
		".page-loading-indicator",
		"#tools-panel",
		".doc-cookie-banner",
		"awsdocs-copyright",
		"awsdocs-thumb-feedback",
	}

	for _, selector := range navSelectors {
		mainContent.Find(selector).Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
	}

	// Convert cleaned HTML to string
	cleanedHTML, err := mainContent.Html()
	if err != nil {
		return fmt.Sprintf("<e>Error extracting cleaned HTML: %v</e>", err)
	}

	// Convert to Markdown using html-to-markdown
	converter := htm.NewConverter("", true, nil)

	// (Optional) Remove specific tags - not built-in like markdownify's `strip`, but you can preprocess if needed
	markdown, err := converter.ConvertString(cleanedHTML)
	if err != nil || strings.TrimSpace(markdown) == "" {
		// todo - fix all of these and throw errors
		return "<e>Page failed to be simplified from HTML</e>"
	}

	return markdown
}
