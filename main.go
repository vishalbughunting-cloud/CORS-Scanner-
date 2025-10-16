package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type CORSTestResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Error         string            `json:"error,omitempty"`
	HasCORS       bool              `json:"has_cors"`
	CORSHeaders   []string          `json:"cors_headers"`
	Timestamp     string            `json:"timestamp"`
}

type Config struct {
	Method      string
	Headers     map[string]string
	Timeout     time.Duration
	Concurrency int
}

func main() {
	var (
		urlInput    = flag.String("url", "", "Single URL to test")
		fileInput   = flag.String("file", "", "File containing URLs (one per line)")
		outputFile  = flag.String("output", "cors_results.txt", "Output file for results")
		method      = flag.String("method", "GET", "HTTP method to use")
		concurrency = flag.Int("concurrency", 5, "Number of concurrent requests")
		timeout     = flag.Int("timeout", 10, "Request timeout in seconds")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("CORS Tool v1.0.0")
		return
	}

	if *urlInput == "" && *fileInput == "" {
		fmt.Println("Error: Either -url or -file must be specified")
		fmt.Println("Usage: cors-tool -url <URL> | -file <filename> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := Config{
		Method:      *method,
		Timeout:     time.Duration(*timeout) * time.Second,
		Concurrency: *concurrency,
		Headers: map[string]string{
			"User-Agent": "CORS-Testing-Tool/1.0",
		},
	}

	var results []CORSTestResult

	if *urlInput != "" {
		if *verbose {
			log.Printf("Testing single URL: %s", *urlInput)
		}
		result := testCORS(*urlInput, config)
		results = append(results, result)
	} else {
		if *verbose {
			log.Printf("Testing URLs from file: %s", *fileInput)
		}
		results = testBulkCORS(*fileInput, config)
	}

	if err := saveResults(results, *outputFile); err != nil {
		log.Fatalf("Error saving results: %v", err)
	}

	if *verbose {
		log.Printf("Results saved to: %s", *outputFile)
	}

	printSummary(results)
}

func testCORS(targetURL string, config Config) CORSTestResult {
	result := CORSTestResult{
		URL:         targetURL,
		Headers:     make(map[string]string),
		Timestamp:   time.Now().Format(time.RFC3339),
		CORSHeaders: []string{},
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	req, err := http.NewRequest(config.Method, targetURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Error creating request: %v", err)
		return result
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	parsedURL, err := url.Parse(targetURL)
	if err == nil {
		origin := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		req.Header.Set("Origin", origin+"-test.cors.com")
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	for key, values := range resp.Header {
		headerValue := strings.Join(values, ", ")
		result.Headers[key] = headerValue
		
		keyLower := strings.ToLower(key)
		if isCORSHeader(keyLower) {
			result.HasCORS = true
			result.CORSHeaders = append(result.CORSHeaders, key)
		}
	}

	return result
}

func testBulkCORS(filename string, config Config) []CORSTestResult {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" && isValidURL(url) {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	if len(urls) == 0 {
		log.Fatal("No valid URLs found in file")
	}

	var results []CORSTestResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, config.Concurrency)

	for _, url := range urls {
		wg.Add(1)
		go func(targetURL string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := testCORS(targetURL, config)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	return results
}

func saveResults(results []CORSTestResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, result := range results {
		// FIXED: Use strings.Repeat instead of incorrect syntax
		writer.WriteString(strings.Repeat("=", 80) + "\n")
		writer.WriteString(fmt.Sprintf("URL: %s\n", result.URL))
		writer.WriteString(fmt.Sprintf("Timestamp: %s\n", result.Timestamp))
		writer.WriteString(fmt.Sprintf("Status Code: %d\n", result.StatusCode))
		writer.WriteString(fmt.Sprintf("Has CORS: %t\n", result.HasCORS))
		
		if result.Error != "" {
			writer.WriteString(fmt.Sprintf("Error: %s\n", result.Error))
		}

		if len(result.CORSHeaders) > 0 {
			writer.WriteString("CORS Headers Found:\n")
			for _, header := range result.CORSHeaders {
				writer.WriteString(fmt.Sprintf("  - %s: %s\n", header, result.Headers[header]))
			}
		}

		writer.WriteString("All Headers:\n")
		for key, value := range result.Headers {
			writer.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		writer.WriteString("\n")
	}

	return nil
}

func printSummary(results []CORSTestResult) {
	total := len(results)
	successful := 0
	withCORS := 0

	for _, result := range results {
		if result.Error == "" {
			successful++
		}
		if result.HasCORS {
			withCORS++
		}
	}

	fmt.Printf("\n=== CORS TEST SUMMARY ===\n")
	fmt.Printf("Total URLs tested: %d\n", total)
	fmt.Printf("Successful requests: %d\n", successful)
	fmt.Printf("URLs with CORS headers: %d\n", withCORS)
	
	if total > 0 {
		fmt.Printf("Success rate: %.2f%%\n", float64(successful)/float64(total)*100)
	} else {
		fmt.Printf("Success rate: 0%%\n")
	}
}

func isCORSHeader(header string) bool {
	corsHeaders := []string{
		"access-control-allow-origin",
		"access-control-allow-methods",
		"access-control-allow-headers",
		"access-control-allow-credentials",
		"access-control-expose-headers",
		"access-control-max-age",
	}

	for _, corsHeader := range corsHeaders {
		if strings.Contains(header, corsHeader) {
			return true
		}
	}
	return false
}

func isValidURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}