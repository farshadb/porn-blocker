package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// implement a set-like structure and check for duplicates
type LinkSet map[string]bool

var blockList = NewLink()
var internalList = NewLink()
var hosts = "hosts.txt"

func NewLink() LinkSet {
	return make(LinkSet)
}

func (link LinkSet) Add(item string) {
	addHost(getHost(item))
	// Normalize before parsing
	lowerCaseURL := strings.ToLower(item)
	// Parse the URL
	parsedURL, err := url.Parse(lowerCaseURL)
	if err != nil {
		return
	}
	// Normalize after parsing
	canonicalURL, _ := normalizeURL(parsedURL)
	// Extract domain
	normalizedDomain := normalizeString(canonicalURL)
	domain := getDomain(normalizedDomain)

	if !blockList.Contains(domain) {
		blockList[domain] = true
	}
}

func (link LinkSet) Contains(item string) bool {
	return blockList[item]
}

func HandleError(err error) error {
	if err != nil {
		return err
	}

	return nil
}

func normalizeString(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = removeAccents(s)
	return norm.NFC.String(s)
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)
	return s
}

func normalizeURL(parsedURL *url.URL) (string, error) {
	if parsedURL == nil {
		return "", fmt.Errorf("parsedURL is nil")
	}
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	parsedURL.Path = url.PathEscape(parsedURL.Path)
	parsedURL.RawQuery = parsedURL.Query().Encode()
	//parsedURL.Fragment = ""

	return parsedURL.String(), nil
}
func isValidLink(link string) bool {

	_, err := url.ParseRequestURI(link)
	HandleError(err)

	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func getDomain(newLink string) string {
	//regular expression pattern to capture the main domain
	re := regexp.MustCompile(`(?:https?://)?(?:www\.)?([^/?&]+)`)
	match := re.FindStringSubmatch(newLink)

	if len(match) > 1 {
		domain := match[1]

		// Remove subdomains by splitting on "." and taking the last two parts
		parts := strings.Split(domain, ".")
		if len(parts) > 2 {
			domain = parts[len(parts)-2] + "." + parts[len(parts)-1]
		}

		// Replace "?" with "/" in the domain
		domain = strings.Replace(domain, "?", "/", -1)
		return domain
	}
	return ""
}

func getHost(link string) string {
	u, err := url.ParseRequestURI(link)
	if err != nil {
		return ""
	}

	parts := strings.Split(u.Host, ".")
	if len(parts) > 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return u.Host
}

func addHost(site string) error {
	file, err := os.OpenFile(hosts, os.O_RDWR|os.O_CREATE, 0755)
	HandleError(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == site {
			return nil
		}
	}

	if _, err := file.WriteString(site + "\n"); err != nil {
		return err
	}
	return nil
}

func removeSites(sitesToRemove []string) error {
	siteList := "blockListCSV.csv"
	file, err := os.Open(siteList)
	HandleError(err)
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	HandleError(err)

	file.Close()

	file, err = os.Create(siteList)
	HandleError(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, line := range lines {
		site := line[0]
		if !isInList(sitesToRemove, site) {
			writer.Write(line)
		}
	}

	fmt.Println("Sites removed from the file.")
	return nil
}

func isInList(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func readLines(pathOrURL string) ([]string, error) {
	_, err := url.ParseRequestURI(pathOrURL)
	fmt.Println(url.ParseRequestURI(pathOrURL))
	if err != nil {
		file, err := os.Open(pathOrURL)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.Trim(line, "\"")     // Remove double quotes
			line = strings.TrimRight(line, ",") // Remove trailing comma
			lines = append(lines, line)
		}
		return lines, scanner.Err()
	} else {
		return []string{string(pathOrURL)}, nil
	}
}

func findURLs(fileNameOrURL string) {

	var externalLinks2 *os.File
	var err error
	var targetList []string
	targetList, _ = readLines(fileNameOrURL)
	fmt.Println(targetList)
	sitesToRemove := []string{""}

	if !isValidLink(fileNameOrURL) {

		externalLinks2, err = os.Create(fileNameOrURL + "_external_links.csv")
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
		}
		defer externalLinks2.Close()
	} else {
		u, _ := url.Parse(fileNameOrURL)
		externalLinks2, err = os.Create(u.Host + "_external_links.csv")
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
		}
		defer externalLinks2.Close()
	}

	for i, list := range targetList {
		fmt.Println(i)
		doc, err := goquery.NewDocument(list)
		if err != nil {
			log.Fatal(err)
		}
		u, _ := url.Parse(list)
		// TODO: in some cases: does not regonize links link this one "https://trends.netcraft.com/topsites"
		// TODO: make this more flexible to find any kind of links in a page
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if isValidLink(href) {
				linkURL, _ := url.Parse(href)
				if linkURL == nil || (linkURL.Host != "" && linkURL.Host != u.Host) {
					blockList.Add(href)
					fmt.Println(href)
				} else if linkURL == nil || (linkURL.Host != "" && linkURL.Host == u.Host) {
					internalList.Add(href)
				}
			}
		})
	}

	if err := removeSites(sitesToRemove); err != nil {
		fmt.Println("Error:", err)
	}
	//Extract keys from the map.
	keys := make([]string, 0, len(blockList))
	for k := range blockList {
		keys = append(keys, k)
	}
	// Sort keys slice based on the length of the key.
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) < len(keys[j])
	})
	for _, link := range keys {
		// Write the link to the CSV file.
		fmt.Fprintln(externalLinks2, link)

	}
}

func main() {
	filename := "https://twitter.com/i/bookmarks"
	findURLs(filename)

}
