package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GrandOichii/mtgsdk"
)

var (
	namesPath     string
	outPath       string
	synergyThresh int
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
	result := make(map[string]map[string]int, len(cnames))
	// save the cards
	for _, cname := range cnames {
		cards, err := mtgsdk.GetCards(map[string]string{mtgsdk.CardNameKey: cname})
		checkErr(err)
		recc, err := cards[0].GetReccomendations(synergyThresh)
		checkErr(err)
		result[cname] = recc
	}
	data, err := json.MarshalIndent(result, "", "\t")
	checkErr(err)
	os.WriteFile(outPath, data, 0755)
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
