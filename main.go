package main

import (
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gocolly/colly"
)

// Album stores information on album found on main page
type Album struct {
	Name          string
	URL           string
	MusicDirector string
	Starring      []string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getDownloadURL(pageURL string) string {
	urlParts, _ := url.Parse(pageURL)
	urlParts.Path = path.Join(urlParts.Path, "download-4.ashx")
	return urlParts.String()
}

func downloadZip(zipURL string) {
	log.Printf("Downloading url - %s\n", zipURL)
	resp, err := http.Get(zipURL)
	check(err)
	var zipFilename string
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		check(err)
		zipFilename = params["filename"]
		if zipFilename == "" {
			urlParts, err := url.Parse(zipURL)
			check(err)
			urlParts.Query().Get("Token")
		}
	}
	log.Printf("Downloading to file : %s", zipFilename)
	zipFile, err := os.Create(zipFilename)
	check(err)
	defer zipFile.Close()
	_, err = io.Copy(zipFile, resp.Body)
	check(err)
}

func main() {
	albums := []Album{}
	mainPageCollector := colly.NewCollector(
		colly.AllowedDomains("www.starmusiq.fun"),
		// colly.CacheDir("./startmusiq_cache"),
	)
	detailsPageCollector := mainPageCollector.Clone()
	mainPageCollector.OnHTML("div.panel-body div.col-xs-6.col-sm-6.col-md-3", func(div *colly.HTMLElement) {
		// log.Println("Found a album ", div.ChildAttr("h5 a", "title"))
		album := Album{
			Name:          strings.TrimSpace(strings.Split(div.ChildAttr("h5 a", "title"), " - ")[0]),
			URL:           div.Request.AbsoluteURL(div.ChildAttr("h5 a", "href")),
			MusicDirector: strings.TrimSpace(div.ChildAttr("div.small a", "title")),
			Starring:      strings.Split(div.ChildAttr("div span", "title"), ", "),
		}
		albums = append(albums, album)
		detailsPageCollector.Visit(album.URL)
	})

	detailsPageCollector.OnHTML(`div.panel-body[itemprop="review"]`, func(div *colly.HTMLElement) {
		log.Println("## Main URL : ", div.Request.URL)
		downloadURLs := []string{"", "", ""}
		div.ForEach("p a[style]", func(_ int, a *colly.HTMLElement) {
			switch text := strings.ToLower(a.Text); {
			case strings.Contains(text, "160kbps"):
				downloadURLs[1] = a.Attr("href")
			case strings.Contains(text, "320kbps"):
				downloadURLs[0] = a.Attr("href")
			default:
				downloadURLs[2] = a.Attr("href")
			}
		})
		for _, downloadURL := range downloadURLs {
			if downloadURL != "" {
				downloadZip(getDownloadURL(downloadURL))
				break
			}
		}
	})
	mainPageCollector.OnHTML(`div.panel-body a[aria-label="Next"]`, func(a *colly.HTMLElement) {
		url := a.Request.AbsoluteURL(a.Attr("href"))
		// log.Println("Found a next page", url)
		mainPageCollector.Visit(url)

	})
	mainPageCollector.Visit("https://www.starmusiq.fun/composers/collection-of-karthik%20raja-starmusiq.html")
	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "  ")
	// enc.Encode(albums)
	// for _, album := range albums {
	// 	fmt.Printf("- %s (%s) %v\n", album.Name, album.MusicDirector, album.Starring)
	// }
}
