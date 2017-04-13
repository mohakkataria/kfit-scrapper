package retriever

import (
	"github.com/PuerkitoBio/goquery"
	//"fmt"
)

// Collection holds all parsed links from supplied URL
type Collection []string

// DocumentBuilder is a type abstraction over our injected dependency
type DocumentBuilder func(url string) (*goquery.Document, error)

// Retrieve function parses provided URL for partner links
func RetrievePartnerLinks(url string, newDoc DocumentBuilder) (Collection, error) {
	coll := Collection{}

	doc, err := newDoc(url)
	if err != nil {
		return Collection{}, err
	}

	doc.Find(".each-card").Each(func(i int, s *goquery.Selection) {	
		s.Find(".card-details").Children().Find("a").Each(func(i int, ss *goquery.Selection) {
			if v, exists := ss.Attr("href"); exists {
				coll = append(coll, v)
			}
		})
	})

	return coll, nil
}