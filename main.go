package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL = "https://ipinfo.io/ips/"
)

var (
	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_3_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 11; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	}
	silentMode = false
)

// logMessage checks silentMode and prints messages only if it's not silent
func logMessage(message string) {
	if !silentMode {
		fmt.Println(message)
	}
}

func randomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

func fetchAndCacheHTML(ip string) (string, error) {
	logMessage("[INFO] Fetching page from the web...")
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", baseURL, ip), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", randomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch IP info for %s: %s", ip, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	cacheFileName := fmt.Sprintf("cache_%s.html", strings.ReplaceAll(ip, "/", "_"))
	err = ioutil.WriteFile(cacheFileName, body, 0644)
	if err != nil {
		return "", err
	}

	logMessage(fmt.Sprintf("[INFO] Page cached successfully: %s", cacheFileName))
	return cacheFileName, nil
}

func loadCachedHTML(fileName string) (*goquery.Document, error) {
	logMessage(fmt.Sprintf("[INFO] Loading cached page: %s...", fileName))
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, err
	}

	logMessage("[INFO] Cached page loaded successfully.")
	return doc, nil
}

func isValidHostname(text string) bool {
	validDomainPattern := `^(([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[A-Za-z]{2,})$`
	matched, err := regexp.MatchString(validDomainPattern, text)
	if err != nil {
		return false
	}
	return matched
}

func extractHostnames(doc *goquery.Document) []string {
	var hostnames []string
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		column := s.Find("td:nth-child(2)")
		text := strings.TrimSpace(column.Text())
		if text != "" && isValidHostname(text) {
			hostnames = append(hostnames, text)
		}
	})
	return hostnames
}

func handleCIDR(cidr string) {
	logMessage(fmt.Sprintf("[INFO] Processing CIDR: %s", cidr))
	cacheFileName := fmt.Sprintf("cache_%s.html", strings.ReplaceAll(cidr, "/", "_"))

	if _, err := os.Stat(cacheFileName); os.IsNotExist(err) {
		newCacheFile, err := fetchAndCacheHTML(cidr)
		if err != nil {
			logMessage(fmt.Sprintf("[ERROR] Failed to fetch and cache page for %s: %v", cidr, err))
			return
		}
		cacheFileName = newCacheFile
	}

	doc, err := loadCachedHTML(cacheFileName)
	if err != nil {
		logMessage(fmt.Sprintf("[ERROR] Failed to load cached page: %v", err))
		return
	}

	hostnames := extractHostnames(doc)
	if len(hostnames) == 0 {
		logMessage("[INFO] No hostnames found.")
	} else {
		for _, hostname := range hostnames {
			fmt.Println(hostname)
		}
	}

	// Remove the cache file after processing
	err = os.Remove(cacheFileName)
	if err != nil {
		logMessage(fmt.Sprintf("[ERROR] Failed to delete cache file %s: %v", cacheFileName, err))
	} else {
		logMessage(fmt.Sprintf("[INFO] Cache file %s deleted successfully.", cacheFileName))
	}
}

func handleCIDRList(filePath string) {
	logMessage(fmt.Sprintf("[INFO] Processing CIDR list from file: %s", filePath))
	file, err := os.Open(filePath)
	if err != nil {
		logMessage(fmt.Sprintf("[ERROR] Failed to open file: %v", err))
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cidr := strings.TrimSpace(scanner.Text())
		if cidr != "" {
			handleCIDR(cidr)
		}
	}

	if err := scanner.Err(); err != nil {
		logMessage(fmt.Sprintf("[ERROR] Error reading file: %v", err))
	}
}

func handleStdin() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		ip := input.Text()
		handleCIDR(ip)
	}

	if err := input.Err(); err != nil {
		logMessage(fmt.Sprintf("[ERROR] Error reading input: %v", err))
	}
}

func main() {
	cidrFlag := flag.String("r", "", "CIDR range to process (e.g., 127.0.0.0/24)")
	listFlag := flag.String("l", "", "Path to a file containing a list of CIDR ranges")
	silentFlag := flag.Bool("silent", false, "Run in silent mode, output only results")
	flag.Parse()

	silentMode = *silentFlag

	if *cidrFlag != "" {
		handleCIDR(*cidrFlag)
	} else if *listFlag != "" {
		handleCIDRList(*listFlag)
	} else {
		handleStdin()
	}
}
