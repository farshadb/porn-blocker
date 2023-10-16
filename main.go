package main

import (
	"bufio"
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
	parsedURL.Fragment = ""

	return parsedURL.String(), nil
}
func isValidLink(link string) bool {

	_, err := url.ParseRequestURI(link)
	if err != nil {
		return false
	}

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
	if err != nil {
		return err
	}
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

func main() {
	targetWebsites := []string{
		"https://pornsites.xxx/",
		"https://www.premiumpornlist.com/",
		"https://area51.to/en/",
		"https://pornwhitelist.com/",
		"https://www.nu-bay.com/categories",
		"http://bigpornlist.com/",
		"https://www.pornpics.com/",
		"https://adultspy.com/porn-lists/",
		"https://toplist18.com/",
		"https://allpornsites.net/",
		"https://thepornbin.com/",
		"https://listofporn.com/",
		"https://www.lindylist.org/",
		"https://freyalist.com/",
		"https://jennylist.xyz/category/list",
		"https://orgasmlist.com/",
		"https://abellalist.com/",
		"https://www.youpornlist.com/",
		"https://pornmate.com/",
		"https://bestlistofporn.com/",
		"https://www.iwantporn.net/",
		"https://porngeek.com/",
		"https://tubepornlist.com/",
		"https://pornlist18.com/",
		"https://www.tblop.com/",
		"https://thesexlist.com/",
		"https://reachporn.com/",
		"https://fivestarpornsites.com/",
		"https://www.primepornlist.com/",
		"https://bestpornsites.org/",
		"https://getpornlist.com/",
		"https://mypornbible.com/",
		"https://darkpornlist.com/",
		"https://onepornlist.com/",
		"https://www.pornlist.tv/",
		"http://abellalist.com/",
		"https://nichepornsites.com/the-50-best-free-porn-sites/",
		"https://www.elephantlist.com/",
		"https://mygaysites.com/",
		"https://pornlist.co/",
		"https://www.mypornlist.net/",
		"https://www.thepornlist.net/",
	}
	blockListCSV, err := os.Create("blockListCSV.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer blockListCSV.Close()

	blockListText, err := os.Create("blockListText.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer blockListText.Close()

	for i, list := range targetWebsites {
		doc, err := goquery.NewDocument(list)
		fmt.Println(i, " ", list)
		if err != nil {
			log.Fatal(err)
		}

		u, _ := url.Parse(list)

		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if isValidLink(href) {
				linkURL, _ := url.Parse(href)
				if linkURL == nil || (linkURL.Host != "" && linkURL.Host != u.Host) {
					blockList.Add(href)
					fmt.Fprintln(blockListText, href)
				}
			}
		})
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
		fmt.Fprintln(blockListCSV, link)

		// Write the link to the TXT file.
		fmt.Fprintln(blockListText, link)
	}
}
