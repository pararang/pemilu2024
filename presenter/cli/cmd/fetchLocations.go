// nolint:typecheck
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

		fileName := fmt.Sprintf("indonesia_location_%s.%s", time.Now().Format("20060102-150405"), fileType)

		if fileType == "json" {
			jsonData, err := json.Marshal(locations)
			if err != nil {
				log.Fatal(err)
			}

			err = os.WriteFile(fileName, jsonData, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

		if fileType == "csv" {
			file, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			if err := writer.Write([]string{"ID", "Code", "Nama", "Level", "ParentID"}); err != nil {
				log.Fatal(err)
			}

			for iProv := 0; iProv < len(locations); iProv++ {
				if err := writer.Write([]string{
					strconv.Itoa(int(locations[iProv].ID)),
					locations[iProv].Code,
					locations[iProv].Name,
					strconv.Itoa(int(locations[iProv].Level)),
					"0",
				}); err != nil {
					log.Fatal(err)
				}

				for iCity := 0; iCity < len(locations[iProv].Cities); iCity++ {
					if err := writer.Write([]string{
						strconv.Itoa(int(locations[iProv].Cities[iCity].ID)),
						locations[iProv].Cities[iCity].Code,
						locations[iProv].Cities[iCity].Name,
						strconv.Itoa(int(locations[iProv].Cities[iCity].Level)),
						strconv.Itoa(int(locations[iProv].ID)),
					}); err != nil {
						log.Fatal(err)
					}

					for iDist := 0; iDist < len(locations[iProv].Cities[iCity].Districts); iDist++ {
						if err := writer.Write([]string{
							strconv.Itoa(int(locations[iProv].Cities[iCity].Districts[iDist].ID)),
							locations[iProv].Cities[iCity].Districts[iDist].Code,
							locations[iProv].Cities[iCity].Districts[iDist].Name,
							strconv.Itoa(int(locations[iProv].Cities[iCity].Districts[iDist].Level)),
							strconv.Itoa(int(locations[iProv].Cities[iCity].ID)),
						}); err != nil {
							log.Fatal(err)
						}
						
						for iSubd := 0; iSubd < len(locations[iProv].Cities[iCity].Districts[iDist].Subdistrict); iSubd++ {
							if err := writer.Write([]string{
								strconv.Itoa(int(locations[iProv].Cities[iCity].Districts[iDist].Subdistrict[iSubd].ID)),
								locations[iProv].Cities[iCity].Districts[iDist].Subdistrict[iSubd].Code,
								locations[iProv].Cities[iCity].Districts[iDist].Subdistrict[iSubd].Name,
								strconv.Itoa(int(locations[iProv].Cities[iCity].Districts[iDist].Subdistrict[iSubd].Level)),
								strconv.Itoa(int(locations[iProv].Cities[iCity].Districts[iDist].ID)),
							}); err != nil {
								log.Fatal(err)
							}
						}
					}
				}
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
