package main

import (
	"encoding/json"
	"log"
	"os"
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

func main() {
	albums := []Album{}
	mainPageCollector := colly.NewCollector(
		colly.AllowedDomains("www.starmusiq.fun"),
		colly.CacheDir("./startmusiq_cache"),
	)
	//detailsPageCollector := mainPageCollector.Clone()
	mainPageCollector.OnHTML("#featured_albums div.col-xs-6.col-sm-6.col-md-3", func(div *colly.HTMLElement) {
		log.Println("Found a album ", div.ChildAttr("h5 a", "title"))
		album := Album{
			Name:          strings.TrimSpace(strings.Split(div.ChildAttr("h5 a", "title"), " - ")[0]),
			URL:           div.Request.AbsoluteURL(div.ChildAttr("h5 a", "href")),
			MusicDirector: strings.TrimSpace(div.ChildAttr("div.small a", "title")),
			Starring:      strings.Split(div.ChildAttr("div span", "title"), ", "),
		}
		albums = append(albums, album)
	})
	mainPageCollector.OnHTML(`#FeatPaginate a[aria-label="Next"]`, func(a *colly.HTMLElement) {
		url := a.Request.AbsoluteURL(a.Attr("href"))
		log.Println("Found a next page", url)
		mainPageCollector.Visit(url)

	})
	mainPageCollector.Visit("https://www.starmusiq.fun/home/")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(albums)
}
