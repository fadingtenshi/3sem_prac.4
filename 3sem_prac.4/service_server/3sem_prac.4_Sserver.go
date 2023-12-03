package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type StatsRequest struct {
	ID          int            `json:"recordId"`
	PID         int            `json:"parent_id"`
	RedirectURL string         `json:"redirect_url"`
	IPAddress   string         `json:"ip_address"`
	Timestamp   string         `json:"timestamp"`
	Count       int            `json:"redirect_count"`
	TimeMomemt  map[string]int `json:"time_moment"`
}

type ReportRequest struct {
	DimensionsOrder []string `json:"dimensionsOrder"`
}

type TimeInterval struct {
	Total int            `json:"Всего"`
	URL   map[string]int `json:"URL"`
}

type IPReport struct {
	Intervals map[string]TimeInterval `json:"Intervals"`
}

type FinalReport map[string]IPReport

const (
	PORT = 6380
)

func writeToFile(stats StatsRequest) {

	_, err := os.Stat("stats.json")

	var file []byte
	var existData []StatsRequest
	latch1 := false
	latch2 := false
	var PID_temp int

	if os.IsNotExist(err) {

		os.Create("stats.json")

		var main StatsRequest

		main.ID = 1
		main.PID = 0
		main.RedirectURL = stats.RedirectURL
		main.Timestamp = "0"
		main.Count++
		main.IPAddress = "0"
		main.TimeMomemt = nil

		stats.ID = 2
		stats.PID = main.ID
		stats.Count = 1
		stats.TimeMomemt = make(map[string]int)
		stats.TimeMomemt[stats.Timestamp] = 1

		existData = append(existData, main, stats)

		newData, err := json.MarshalIndent(existData, "", "    ")
		if err != nil {

			fmt.Println("Error marshalling data to JSON: ", err)
			return

		}

		err = os.WriteFile("stats.json", newData, 0644)
		if err != nil {
			fmt.Println("Error writing data to file:", err)
			return
		}

		fmt.Println("Stats written to stats.json:", stats)

	} else {

		file, err = os.ReadFile("stats.json")
		if err != nil {

			fmt.Println("Error opening or creating file: ", err)
			return

		}

		err = json.Unmarshal(file, &existData)
		if err != nil {

			fmt.Println("Error unmarshalling existing data: ", err)
			return

		}

		for i := 0; i < len(existData); i++ {
			if existData[i].PID == 0 {
				if existData[i].RedirectURL == stats.RedirectURL {

					existData[i].Count++
					latch1 = true
					PID_temp = existData[i].ID

				}
			} else {

				if existData[i].RedirectURL == stats.RedirectURL {

					latch1 = true

					if existData[i].IPAddress == stats.IPAddress {

						existData[i].Count++
						existData[i].TimeMomemt[stats.Timestamp]++
						latch2 = true

					}
				}

			}
		}

		if !latch1 && !latch2 {

			var main StatsRequest

			main.ID = len(existData) + 1
			main.PID = 0
			main.RedirectURL = stats.RedirectURL
			main.Timestamp = "0"
			main.Count++
			main.IPAddress = "0"
			main.TimeMomemt = nil

			stats.ID = len(existData) + 2
			stats.PID = main.ID
			stats.Count = 1
			stats.TimeMomemt = make(map[string]int)
			stats.TimeMomemt[stats.Timestamp]++

			existData = append(existData, main, stats)

		} else if latch1 && !latch2 {

			stats.Count++
			stats.ID = len(existData) + 1
			stats.PID = PID_temp
			stats.TimeMomemt = make(map[string]int)
			stats.TimeMomemt[stats.Timestamp]++
			existData = append(existData, stats)

		}

		newData, err := json.MarshalIndent(existData, "", "    ")
		if err != nil {

			fmt.Println("Error marshalling data to JSON: ", err)
			return

		}

		err = os.WriteFile("stats.json", newData, 0644)
		if err != nil {
			fmt.Println("Error writing data to file:", err)
			return
		}

		fmt.Println("Stats written to stats.json:", stats)

	}

}

func getInfo(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		var stats StatsRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&stats)
		if err != nil {

			http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
			return

		}
		// INFO
		writeToFile(stats)

		w.WriteHeader(http.StatusOK)

	} else {

		fmt.Fprintf(w, "Invalid method")
		return

	}

}

func getNextMin(stat StatsRequest) string {
	parsedTime, err := time.Parse("15:04", stat.Timestamp)
	if err != nil {
		return ""
	}
	newTime := parsedTime.Add(1 * time.Minute)
	newTimeString := newTime.Format("15:04")
	return newTimeString
}

func getReport(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var repOrder ReportRequest
		var existData []StatsRequest

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&repOrder)
		if err != nil {
			http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
			return
		}

		_, err = os.Stat("stats.json")
		if os.IsNotExist(err) {
			fmt.Println("file's empty")
			return
		} else {
			file, err := os.ReadFile("stats.json")
			if err != nil {
				fmt.Println("Error opening or creating file: ", err)
				return
			}

			err = json.Unmarshal(file, &existData)
			if err != nil {
				fmt.Println("Error unmarshalling existing data: ", err)
				return
			}

			finalReport := make(FinalReport)

			for _, stat := range existData {
				if stat.IPAddress == "0" {
					continue
				}
				ipReport, ok := finalReport[stat.IPAddress]
				if !ok {
					ipReport = IPReport{Intervals: make(map[string]TimeInterval)}
				}

				timeInterval, ok := ipReport.Intervals[stat.Timestamp+" - "+getNextMin(stat)]
				if !ok {
					timeInterval = TimeInterval{URL: make(map[string]int)}
				}

				timeInterval.Total += stat.Count
				timeInterval.URL[stat.RedirectURL] += stat.Count

				ipReport.Intervals[stat.Timestamp+" - "+getNextMin(stat)] = timeInterval
				finalReport[stat.IPAddress] = ipReport
			}

			reportJSON, err := json.MarshalIndent(finalReport, "", "    ")
			if err != nil {
				fmt.Println("Error marshalling report data: ", err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(reportJSON)

		}
	} else {
		fmt.Fprintf(w, "Invalid method")
		return
	}
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", getInfo)
	mux.HandleFunc("/report", getReport)

	fmt.Println("HTTP Server listening on " + strconv.Itoa(PORT))
	err := http.ListenAndServe(":"+strconv.Itoa(PORT), mux)
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
	}

}
