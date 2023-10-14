package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var blockList = make(map[string]bool)

var searchList = make(map[string]bool)

// func addToTargetList(newLink string) {
// 	if searchList[newLink] {
// 		n++
// 		fmt.Println(newLink)

// 		return
// 	} else {
// 		searchList[newLink] = true
// 	}
// }

func addToBlockList(newLink string) {
	if blockList[newLink] {
		return
	} else {
		blockList[newLink] = true
	}
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

	for _, list := range targetWebsites {

		doc, err := goquery.NewDocument(list)
		if err != nil {
			log.Fatal(err)
		}
		u, _ := url.Parse(list)
		//fmt.Println(u)
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

		// Create a new TXT file.
		finalTxt, err := os.Create("links.txt")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer finalTxt.Close()

		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				if blockList[href] {
					linkURL, _ := url.Parse(href)
					if linkURL != nil && (linkURL.Host == "" || linkURL.Host == u.Host) {
						fmt.Fprintln(internalFile, href)
					} else {
						addToBlockList(href)
						fmt.Fprintln(finalTxt, href)
					}
				}
			}
		})
	}

	finalBlockListCSV, err := os.Create("finalBlockListCSV.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer finalBlockListCSV.Close()

	finalWebsiteTXT, err := os.Create("finalWebsiteTXT.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer finalWebsiteTXT.Close()

	for link := range blockList {
		// Remove "http://" from the link.
		link = strings.TrimPrefix(link, "http://")

		// Write the link to the CSV file.
		fmt.Fprintln(finalBlockListCSV, link)

		// Write the link to the TXT file.
		fmt.Fprintln(finalWebsiteTXT, link)
	}

}
