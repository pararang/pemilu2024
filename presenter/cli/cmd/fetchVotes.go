/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

		controller := controller.NewController(kpu.NewSirekap(stdHttpClient))
		votes, err := controller.GetVotesNationwide()
		if err != nil {
			log.Fatal(err)
		}

		localData := struct {
			LocalTimestamp string `json:"local_timestamp"`
			Raw kpu.ResponseDataNationwide `json:"raw_data"`
		} {
			LocalTimestamp: time.Now().UTC().Format(time.RFC3339),
			Raw: votes,
		}

		jsonData, err := json.MarshalIndent(localData, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile("votes_nationwide.json", jsonData, 0644)
		if err != nil {
			log.Fatal(err)
		}
	},
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
