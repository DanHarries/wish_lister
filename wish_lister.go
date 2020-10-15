package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"strconv"
)

type Items struct {
	id        string
	product   string
	price     string
	url       string
	rating    string
	dateAdded string
}

type WishListItem struct {
	person string
	items  []Items
}

func main() {

	var wl WishListItem

	if len(os.Args) < 2 {
		log.Fatalln("Please enter a culture and a wish list id \nFor example: .co.uk 2OABCDE0FGH42")
	}

	// Get all data from person's wish list
	wl = getWishList(os.Args[1], os.Args[2])

	fmt.Println(wl.person)
	for _, item := range wl.items {

		fmt.Printf("Item: %s \nPrice: %1s \nRating: %3s \n%3s\n\n",
			item.product, item.price, item.rating, item.dateAdded)

	}

}

// Extract items from wish list
func getWishList(culture string, wishListId string) WishListItem {

	var wlItems WishListItem
	getCultureUrl := GetCultureUrl(culture)
	getWishListUrl := WishListUrl(wishListId)

	// Instantiate default collector
	c := colly.NewCollector()

	// Call back to extract wish list items
	c.OnHTML("#g-items", func(e *colly.HTMLElement) {

		// We need the count of how many items are in the wish list, this is in a hidden input.
		count, err := strconv.Atoi(e.DOM.Nodes[0].FirstChild.Attr[2].Val)

		if err != nil {
			log.Panicf("Something went wrong %s", err.Error())
		}

		for i := 0; i < count+1; i++ {

			// Get the li of the wish list, this has an index from 2
			liSelector := fmt.Sprintf("#g-items > li:nth-child(%d)", i+2)

			// loop over the list
			e.ForEach(liSelector, func(i int, e *colly.HTMLElement) {

				// Get the item id
				itemId := e.DOM.Nodes[0].Attr[1].Val

				// If item id isn't found then it the rest won't work
				if itemId == "" {
					log.Fatalf("Item id %s not found", itemId)
					return
				}

				// These selectors are suffixed with the item id, so format these strings
				anchorSelector := fmt.Sprintf("#itemImage_%s > a", itemId)

				// Get the image
				url := e.ChildAttr(anchorSelector, "href")
				url = fmt.Sprintf("%5s%4s", getCultureUrl, url)

				// The title is on the anchor tag so grab that whilst we're at it
				product := e.ChildAttr(anchorSelector, "title")

				// Get the price, this is buried inside a span
				priceSelector := fmt.Sprintf("#itemPrice_%s > span[class='a-offscreen']", itemId)
				price := e.ChildText(priceSelector)

				// Price can be empty if out of stock, discontinued etc.
				if price == "" {
					price = "Price currently unavailable"
				}

				// Get the rating
				ratingSelector := fmt.Sprintf("#review_stars_%s > span", itemId)
				rating := e.ChildText(ratingSelector)

				// Get the date added
				dateAddedSelector := fmt.Sprintf("#itemAddedDate_%s", itemId)
				dateAdded := e.ChildText(dateAddedSelector)

				// Get the values in to the struct

				wlItems.items = append(wlItems.items,
					Items{itemId, product, price, url, rating, dateAdded})

			})

		}

	})

	// Get the person's name
	c.OnHTML("#profile-list-name", func(e *colly.HTMLElement) {

		// Add name to struct
		wlItems.person = e.Text

	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Set up url
	scrapingSite := fmt.Sprintf("%5s%4s", getCultureUrl, getWishListUrl)

	// Start scraping
	err := c.Visit(scrapingSite)

	if err != nil {
		log.Fatalf("Error starting scrape: %s", err.Error())
	}

	return wlItems

}

// This sorts the culture id
func GetCultureUrl(culture string) string {

	return fmt.Sprintf("https://www.amazon%s", culture)

}

// This formats it correctly
func WishListUrl(wishListId string) string {

	return fmt.Sprintf("/hz/wishlist/ls/%s", wishListId)

}
