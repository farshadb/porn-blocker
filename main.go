package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	siteURL := "https://pornsites.xxx/"
	doc, err := goquery.NewDocument(siteURL)
	if err != nil {
		log.Fatal(err)
	}
	u, _ := url.Parse(siteURL)
	siteName := strings.ReplaceAll(u.Host, ".", "_")

	internalFile, err := os.Create(siteName + "_internal_links.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer internalFile.Close()

	externalFile, err := os.Create(siteName + "_external_links.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer externalFile.Close()

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			linkURL, _ := url.Parse(href)
			if linkURL != nil && (linkURL.Host == "" || linkURL.Host == u.Host) {
				fmt.Fprintln(internalFile, href)
			} else {
				fmt.Fprintln(externalFile, href)
			}
		}
	})
}
