package main

import (
	//"encoding/json"
	"fmt"
	"os"
	"net/url"
	"time"
	"strconv"
	"github.com/PuerkitoBio/goquery"
	"github.com/codegangsta/cli"
	r "github.com/mohakkataria/kfit-scraper/retriever"
	s "github.com/mohakkataria/kfit-scraper/scraper"
	w "github.com/mohakkataria/kfit-scraper/writer"
)

const gateway = "https://access.kfit.com/partners?city=kuala-lampur"
const maxPages = 10000

func main() {
	app := cli.NewApp()
	app.Name = "KFit Scraper"
	app.Version = "0.0.1"
	app.Author = "Mohak Kataria"
	app.Usage = "CLI tool for scraping contents of partners from KFit Kuala Lampur page"
	app.Action = process

	app.CommandNotFound = commandNotFound
	app.Run(os.Args)
}

func process(c *cli.Context) {
	
	ch := make(chan r.Collection)
	quit := make(chan int)
	pagesFetched := 0
	go func () {
		fetch := true
		for i := 0; i < maxPages; i++ {
			time.Sleep(500*time.Millisecond)
			if (fetch == false) {
				break
			}
			go func(i int) {
				//fmt.Println(i)
				v := url.Values{}
				v.Set("page", strconv.Itoa(i))
				coll, err := r.RetrievePartnerLinks(gateway+"&"+v.Encode(), goquery.NewDocument)
				if err != nil {
					fmt.Printf("There was an issue retrieving links from the page: %s", err.Error())
					os.Exit(1)
				}
				//fmt.Println(coll)

				if (len(coll) == 0) {
					fetch = false
					quit <- i
				} else {
					pagesFetched++
					ch <- coll
				}
			}(i)
		}
	}()
	i := 0
	shouldContinue := true
	for {
		if (shouldContinue == false && i == pagesFetched) {
			break
		}
		select {
		case x := <- ch:
			i++
			// scrape the page here and retrieve the data from the partner page now
			b := s.Scrape(x)
			// write the data to the file
			w.Write(b.Partners)
			//fmt.Println(b)
		case <-quit:
			shouldContinue = false
		}
	}


	// b, err := json.MarshalIndent(s.Scrape(coll), "", "    ")
	// if err != nil {
	// 	fmt.Printf("There was an issue converting our data into JSON: %s", err.Error())
	// 	os.Exit(1)
	// }

	// fmt.Println(string(b))

	//fmt.Println(coll)
}

func commandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
