package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/GrandOichii/colorwrapper"
	"github.com/GrandOichii/mtgsdk"
)

type PrintFlag struct {
	name  string
	usage string
	use   bool
	print func(cards map[*mtgsdk.Card]int) error
}

var (
	printFlags = []*PrintFlag{
		{
			name:  "cardprint",
			usage: "(optional) Prints the cards to the console as cards",
			use:   false,
			print: func(cards map[*mtgsdk.Card]int) error {
				for card, syn := range cards {
					err := card.CardPrint()
					if err != nil {
						return err
					}
					fmt.Printf("== Synergy: %d ==\n", syn)
				}
				fmt.Println()
				return nil
			},
		},
		{
			name:  "print",
			usage: "(optional) Prints the card names out to the console",
			use:   false,
			print: func(cards map[*mtgsdk.Card]int) error {
				for card, syn := range cards {
					sign := ""
					if syn > 0 {
						sign = "+"
					}
					fmt.Printf("\t%s -- %s%d\n", card.Name, sign, syn)
				}
				fmt.Println()
				return nil
			},
		},
	}
)

var (
	namesPath     string
	outPath       string
	synergyThresh int
	logF          bool
)

func init() {
	flag.StringVar(&namesPath, "src", "", "The path to the file with the commander names")
	flag.StringVar(&outPath, "out", "", "The out path")
	flag.IntVar(&synergyThresh, "synergy", 10, "(optional) The threshold for the synergy")
	flag.BoolVar(&logF, "log", false, "(optional) Log the messages")

	for _, pflag := range printFlags {
		flag.BoolVar(&pflag.use, pflag.name, false, pflag.usage)
	}
}

func main() {
	flag.Parse()
	if !logF {
		log.SetOutput(ioutil.Discard)
	}
	if namesPath == "" {
		flag.PrintDefaults()
		return
	}
	if outPath == "" {
		flag.PrintDefaults()
		return
	}
	// read the commander names
	cnames, err := readCommanderNames(namesPath)
	checkErr(err)
	result := make(map[string]map[*mtgsdk.Card]int, len(cnames))
	// save the cards
	for _, cname := range cnames {
		if cname == "" {
			continue
		}
		cards, err := mtgsdk.GetCards(map[string]string{mtgsdk.CardNameKey: cname})
		commanderName := cards[0].Name
		fmt.Printf("Getting reccomendations for %s with synergy %d...\n", commanderName, synergyThresh)
		checkErr(err)
		if len(cards) == 0 {
			fmt.Printf("No commander with name %s\n", cname)
			return
		}
		ci := 0
		for {
			if cards[ci].IsCreature() && cards[ci].IsLegendary() {
				break
			}
			ci++
			if ci == len(cards) {
				fmt.Printf("No commander with name %s\n", cname)
				return
			}
		}
		recc, err := cards[ci].GetReccomendations(synergyThresh)
		checkErr(err)
		result[commanderName] = recc
		fmt.Printf("Cards for %s loaded!\n", commanderName)
	}
	err = writeToFile(result, outPath)
	checkErr(err)
	fmt.Println("Done!")
	err = printCards(result)
	checkErr(err)
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

func writeToFile(m map[string]map[*mtgsdk.Card]int, outPath string) error {
	result := make(map[string]map[string]int, len(m))
	for key, cm := range m {
		result[key] = make(map[string]int, len(cm))
		for card, syn := range cm {
			result[key][card.Name] = syn
		}
	}
	data, err := json.MarshalIndent(result, "", "\t")
	checkErr(err)
	return os.WriteFile(outPath, data, 0755)
}

func printCards(cm map[string]map[*mtgsdk.Card]int) error {
	for _, pflag := range printFlags {
		if pflag.use {
			err := colorwrapper.Printf("red-black", "%s:\n", pflag.name)
			if err != nil {
				return err
			}
			for cname, cmap := range cm {
				fmt.Printf("Reccomendations for %s:\n", cname)
				err = pflag.print(cmap)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
