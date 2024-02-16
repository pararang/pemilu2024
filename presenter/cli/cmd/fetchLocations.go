/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pararang/pemilu2024/controller"
	"github.com/pararang/pemilu2024/kpu"
	"github.com/spf13/cobra"
)

// fetchLocationsCmd represents the fetchLocations command
var fetchLocationsCmd = &cobra.Command{
	Use:   "fetchLocations",
	Short: "Fetc location and save it to the persistent storage",
	Long:  "Fetc location and save it to the persistent storage",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("fetchLocations called, fileType=%s\n", fileType)

		start := time.Now()
		defer func(start time.Time) {
			log.Printf("done after %s", time.Since(start).String())
		}(start)

		controller := controller.NewController(kpu.NewSirekap(http.DefaultClient))
		locations, err := controller.GetLocations()
		if err != nil {
			log.Fatal(err)
		}

		if fileType == "json" {
			jsonData, err := json.Marshal(locations)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fileName := fmt.Sprintf("indonesia_location_%s.json", time.Now().Format("20060102-150405"))
			err = os.WriteFile(fileName, jsonData, 0644)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}


	},
}

func init() {
	rootCmd.AddCommand(fetchLocationsCmd) //nolint:typecheck

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchLocationsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchLocationsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
