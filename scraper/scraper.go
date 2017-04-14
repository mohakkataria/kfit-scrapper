package scraper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "regexp"
	"strconv"
	"strings"
	"sync"
	r "github.com/mohakkataria/kfit-scraper/retriever"
	"github.com/PuerkitoBio/goquery"
)

// Partner stores details of a single partner
type Partner struct {
	Name       string `csv:"name"`
	City        string `csv:"city"`
	Address 	string `csv:"address"`
	Latitude   float64 `csv:"latitude"`
	Longitude float64 `csv:"longitude"`
	Rating float64 `csv:"rating"`
	Phone string `csv:"phone"`
}

// Result stores details of the scraped partners
type Result struct {
	Partners []Partner `json:"results"`
}

type extendedDocument struct {
	Size     string
	Document *goquery.Document
}

var ch chan Partner
var wg sync.WaitGroup

const host = "https://access.kfit.com"

// Scrape function parses provided URL for product links
func Scrape(urls []string) Result {
	ch = make(chan Partner, len(urls))

	result := Result{}

	for _, url := range urls {
		wg.Add(1)
		go getPartner(host+url)
	}
	wg.Wait()
	close(ch)

	for item := range ch {
		result.Partners = append(result.Partners, item)
	}

	return result
}

func extendDocument(url string) (extendedDocument, error) {
	res, err := http.Get(url)
	if err != nil {
		return extendedDocument{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return extendedDocument{}, err
	}
	size := strconv.Itoa(len(body)/1000) + "kb"

	// Rewind response body so it can be re-read by goquery
	res.Body = ioutil.NopCloser(bytes.NewReader(body))

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return extendedDocument{}, err
	}

	return extendedDocument{size, doc}, nil
}

var getPartner = func(url string) {
	defer wg.Done()

	d, err := extendDocument(url)
	if err != nil {
		fmt.Println(
			fmt.Errorf("Unable to create a new document: %s", err.Error()),
		)
	}

	partner := Partner{}

	rating := d.Document.Find("span .rating").Text()
	partner.Rating,_ = strconv.ParseFloat(rating, 64)
	text := ""
	d.Document.Find("script").Each(func(i int, s *goquery.Selection) {
		if (strings.Contains(s.Text(), "var outlet_details = ")) {
			text = s.Text()
		}
	})

	if (len(text) > 0) {
		//parse and remove remaining data
		firstOpeningBracketIndex := strings.IndexRune(text,'{')
		firstClosingBracketIndex := strings.IndexRune(text,'}')
		partnerDataJsonString := text[firstOpeningBracketIndex:firstClosingBracketIndex+1]
		nameIndex := strings.Index(partnerDataJsonString, "name:")
		addressIndex := strings.Index(partnerDataJsonString, "address:")
		positionIndex := strings.Index(partnerDataJsonString, "position:")
		cityIndex := strings.Index(partnerDataJsonString, "city:")
		iconIndex := strings.Index(text, "icon:")
		partner.Name = strings.Trim(strings.Trim(strings.TrimSpace(partnerDataJsonString[nameIndex+5:addressIndex-1]),","),"'")
		partner.Address = strings.Trim(strings.Trim(strings.TrimSpace(partnerDataJsonString[addressIndex+8:cityIndex-2]),","),"'")
		partner.City = strings.Trim(strings.Trim(strings.TrimSpace(strings.ToTitle(strings.Replace(partnerDataJsonString[cityIndex+5:positionIndex-1], "-"," ", -1))),","),"'")
		positionString := partnerDataJsonString[positionIndex+9:iconIndex]
		positionString = positionString[strings.IndexRune(positionString,'(')+1:strings.IndexRune(positionString,')')]
		positionCoordinates := strings.Split(positionString,",")
		latitude,_ := strconv.ParseFloat(strings.Trim(strings.TrimSpace(positionCoordinates[0]),"'"), 64)
		longitude,_ := strconv.ParseFloat(strings.Trim(strings.TrimSpace(positionCoordinates[1]),"'"), 64)
		partner.Latitude = latitude
		partner.Longitude = longitude
	}
	
	if d.Document.Find(".classtable .emptytable").Length() == 0 {
		if v, exists := d.Document.Find(".classtable .reserve-col .btn").First().Attr("href"); exists {
			partner.Phone,_ = r.RetrievePartnerPhone(host+v, goquery.NewDocument)
		}
	}

	ch <- partner
	
}

func (Partner) GetHeaders() []string{
	return []string{"name", "address", "city", "latitude", "longitude", "rating", "phone"}
}

func (p *Partner) GetSerializedData() []string{
	return []string{p.Name, p.Address, p.City, strconv.FormatFloat(p.Latitude, 'f', 6, 64), strconv.FormatFloat(p.Longitude, 'f', 6, 64), strconv.FormatFloat(p.Rating, 'f', 1, 64), p.Phone}
}