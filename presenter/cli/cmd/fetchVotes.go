/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pararang/pemilu2024/controller"
	"github.com/pararang/pemilu2024/kpu"
	"github.com/spf13/cobra"
)

// fetchVotesCmd represents the fetchVotes command
var fetchVotesCmd = &cobra.Command{
	Use:   "fetchVotes",
	Short: "fetch votes",
	Long:  "fetch votes from KPU and save it to the file(s)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fetchVotes called")

		sirekapClient := kpu.NewSirekap(stdHttpClient)
		controller := controller.NewController(sirekapClient)
		votesPresident, err := controller.GetVotesNationwide()
		if err != nil {
			log.Fatal(err)
		}

		localData := struct {
			LocalTimestamp string                                 `json:"local_timestamp"`
			Raw            kpu.ResponseDataPresidentialNationwide `json:"raw_data"`
		}{
			LocalTimestamp: timeProcessed.Format(time.RFC3339),
			Raw:            votesPresident,
		}

		jsonData, err := json.MarshalIndent(localData, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile("output/votes/votes_nationwide.json", jsonData, 0644)
		if err != nil {
			log.Fatal(err)
		}

		var provinces kpu.Locations
		err = sirekapClient.FetchLocations(&provinces, "0")
		if err != nil {
			log.Fatal(err)
		}

		var mapProvName = make(map[string]string, 0)
		for _, prov := range provinces {
			mapProvName[prov.Code] = prov.Name
		}

		saveVotesPresidential(mapProvName, votesPresident)

		votesLegislative, err := sirekapClient.GetVotesLegislativeNationwide()
		if err != nil {
			log.Fatal(err)
		}
		saveVotesLegislative(mapProvName, votesLegislative)
	},
}

// TODO:return and handle error
func saveVotesPresidential(mapProvName map[string]string, votes kpu.ResponseDataPresidentialNationwide) {
	for code, name := range mapProvName {
		vote, ok := votes.Table[code]
		if ok {
			filename := strings.ReplaceAll(fmt.Sprintf("output/votes/votes_0_%s.csv", strings.ToLower(name)), " ", "_")
			var osFile *os.File
			_, err := os.Stat(filename)
			var isCreate bool
			if os.IsNotExist(err) {
				// File doesn't exist, create it
				osFile, err = os.Create(filename)
				isCreate = true
			} else {
				// File exists, open it in append mode
				osFile, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			}

			if err != nil {
				log.Fatal(err)
			}
			defer osFile.Close()

			osWriter := csv.NewWriter(osFile)
			defer osWriter.Flush()

			if isCreate {
				if err := osWriter.Write([]string{"ts", "amin", "pagi", "gama", "created_at"}); err != nil {
					log.Fatal(err)
				}
			}
			if err := osWriter.Write([]string{
				votes.Ts,
				fmt.Sprintf("%d", *vote.The100025),
				fmt.Sprintf("%d", *vote.The100026),
				fmt.Sprintf("%d", *vote.The100027),
				timeProcessed.Format(time.RFC3339),
			}); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// TODO:return and handle error
func saveVotesLegislative(mapProvName map[string]string, votes kpu.ResponseDataLegislativeNationwide) {
	for code, name := range mapProvName {
		vote, ok := votes.Table[code]
		if ok {
			filename := strings.ReplaceAll(fmt.Sprintf("output/votes/votes_dpr_0_%s.csv", strings.ToLower(name)), " ", "_")
			var osFile *os.File
			_, err := os.Stat(filename)
			var isCreate bool
			if os.IsNotExist(err) {
				// File doesn't exist, create it
				osFile, err = os.Create(filename)
				isCreate = true
			} else {
				// File exists, open it in append mode
				osFile, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			}

			if err != nil {
				log.Fatal(err)
			}
			defer osFile.Close()

			osWriter := csv.NewWriter(osFile)
			defer osWriter.Flush()

			if isCreate {
				if err := osWriter.Write([]string{
					"ts",
					"created_at",
					"PKB",
					"Gerindra",
					"PDI-P",
					"Golkar",
					"Nasdem",
					"Partai Buruh",
					"Gelora",
					"PKS",
					"PKN",
					"Hanura",
					"Garuda",
					"PAN",
					"PBB",
					"Demokrat",
					"PSI",
					"Perindo",
					"PPP",
					"Partai Ummat",
				}); err != nil {
					log.Fatal(err)
				}
			}
			if err := osWriter.Write([]string{
				votes.Ts,
				timeProcessed.Format(time.RFC3339),
				strconv.Itoa(int(vote.The1)),
				strconv.Itoa(int(vote.The2)),
				strconv.Itoa(int(vote.The3)),
				strconv.Itoa(int(vote.The4)),
				strconv.Itoa(int(vote.The5)),
				strconv.Itoa(int(vote.The6)),
				strconv.Itoa(int(vote.The7)),
				strconv.Itoa(int(vote.The8)),
				strconv.Itoa(int(vote.The9)),
				strconv.Itoa(int(vote.The10)),
				strconv.Itoa(int(vote.The11)),
				strconv.Itoa(int(vote.The12)),
				strconv.Itoa(int(vote.The13)),
				strconv.Itoa(int(vote.The14)),
				strconv.Itoa(int(vote.The15)),
				strconv.Itoa(int(vote.The16)),
				strconv.Itoa(int(vote.The17)),
				strconv.Itoa(int(vote.The24)),
			}); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(fetchVotesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchVotesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchVotesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
