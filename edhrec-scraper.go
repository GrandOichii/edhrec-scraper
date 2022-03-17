package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-rod/rod"
)

const (
	edhrecURL = "https://edhrec.com/commanders/%s"

	cardSelector = "div[class^=\"Card_container__\"]"
)

var (
	dchars = []string{
		",",
	}
)

var (
	namesPath     string
	outPath       string
	synergyThresh int

	browser *rod.Browser
)

func init() {
	flag.StringVar(&namesPath, "src", "", "The path to the file with the commander names")
	flag.StringVar(&outPath, "out", "", "The out path")
	flag.IntVar(&synergyThresh, "synergy", 10, "The threshold for the synergy")
}

func main() {
	flag.Parse()
	if namesPath == "" {
		checkErr(flag.ErrHelp)
	}
	if outPath == "" {
		checkErr(flag.ErrHelp)
	}
	// read the commander names
	cnames, err := readCommanderNames(namesPath)
	checkErr(err)
	// parse command line arguments
	err = saveTo(cnames, outPath)
	checkErr(err)
	fmt.Println("Done!")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func readCommanderNames(dir string) ([]string, error) {
	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func commanderURL(cname string) string {
	cname = strings.ToLower(cname)
	for _, dchar := range dchars {
		cname = strings.ReplaceAll(cname, dchar, "")
	}
	cname = strings.ReplaceAll(cname, " ", "-")
	return fmt.Sprintf(edhrecURL, cname)
}

func initBrowser() error {
	browser = rod.New()
	err := browser.Connect()
	if err != nil {
		return err
	}
	fmt.Println("Connected to browser")
	return nil
}

func nav(url string) (*rod.Page, error) {
	page := browser.MustPage()
	waitFunc := page.MustWaitNavigation()
	err := page.Navigate(url)
	if err != nil {
		return nil, err
	}
	if waitFunc != nil {
		waitFunc()
		waitFunc = nil
	}
	fmt.Println("Connected to the page, rendering...")
	fmt.Println("Page rendered!")
	return page, nil
}

func extractNameAndSynergy(text string) (string, int, error) {
	lines := strings.Split(text, "\n")
	s, err := strconv.Atoi(strings.Split(lines[5], "%")[0])
	return lines[3], s, err
}

func saveTo(cnames []string, dir string) error {
	// initialize the browser
	err := initBrowser()
	defer browser.MustClose()
	if err != nil {
		return err
	}

	result := map[string]map[string]int{}

	// convert name to url
	for _, cname := range cnames {

		fmt.Printf("Searching best cards for %s\n", cname)
		url := commanderURL(cname)
		fmt.Printf("Connecting to %s...\n", url)

		// connect to browser
		page, err := nav(url)
		if err != nil {
			return err
		}

		// scrape the elements
		cards := page.MustElements(cardSelector)[1:]
		amount := len(cards)
		fmt.Printf("Found %v cards\n", amount)
		result[cname] = make(map[string]int, amount)
		for _, card := range cards {
			text, err := card.Text()
			if err != nil {
				return err
			}
			name, synergy, err := extractNameAndSynergy(text)
			if err != nil {
				return err
			}
			if synergy > synergyThresh {
				result[cname][name] = synergy
			}
		}
		fmt.Printf("Card stats for %s saved!\n", cname)
		err = page.Close()
		if err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(dir, data, 0755)
	return err
}
