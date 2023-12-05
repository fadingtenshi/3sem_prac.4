package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

type StatsRequest struct {
	ID          int    `json:"recordId"`
	PID         int    `json:"parent_id"`
	RedirectURL string `json:"redirect_url"`
	IPAddress   string `json:"ip_address"`
	Timestamp   string `json:"timestamp"`
	Count       int    `json:"redirect_count"`
}

type StatsReport struct {
	ID           int                     `json:"Id"`
	PID          *int                    `json:"Pid,omitempty"`
	URL          *string                 `json:"URL,omitempty"`
	SourceIP     *string                 `json:"SourceIP,omitempty"`
	TimeInterval *string                 `json:"TimeInterval,omitempty"`
	Count        int                     `json:"Count"`
	Children     map[string]*StatsReport `json:"Children,omitempty"`
}

type ReportRequest struct {
	DimensionsOrder []string `json:"dimensionsOrder"`
}

type Node struct {
	Data StatsRequest `json:"data"`
	Next *Node        `json:"next"`
}

type Queue struct {
	First *Node `json:"first"`
	Last  *Node `json:"last"`
}

func (q *Queue) isEmpty() bool {
	return q.First == nil
}

func (q *Queue) pop() StatsRequest {

	if q.isEmpty() {
		return StatsRequest{}
	}

	value := q.First.Data
	temp := q.First
	q.First = temp.Next

	temp.Next = nil
	temp.Data = StatsRequest{}
	temp = nil

	return value
}

const (
	PORT = 6380
	IP   = "192.168.177.235" // Your IPv4
)

func buildReport(dimensions []string, data []StatsRequest) *StatsReport {
	root := &StatsReport{
		ID:       1,
		URL:      nil,
		SourceIP: nil,
		Count:    0,
		Children: make(map[string]*StatsReport),
	}

	for _, stat := range data {
		currentNode := root
		for _, dimension := range dimensions {
			key := getKey(stat, dimension)
			if currentNode.Children[key] == nil {
				currentNode.Children[key] = &StatsReport{
					ID:           len(currentNode.Children) + 2,
					PID:          &currentNode.ID,
					URL:          nil,
					SourceIP:     nil,
					TimeInterval: nil,
					Count:        1,
					Children:     make(map[string]*StatsReport),
				}
			} else {
				currentNode.Children[key].Count++
			}
			currentNode = currentNode.Children[key]
		}

	}

	for _, child := range root.Children {
		child.Count = 0
		for _, subChild := range child.Children {
			child.Count += subChild.Count
		}
	}

	for _, child := range root.Children {

		root.Count += child.Count

	}

	return root
}

func getKey(stat StatsRequest, dimension string) string {
	switch dimension {
	case "RedirectURL":
		return stat.RedirectURL
	case "IPAddress":
		return stat.IPAddress
	case "Timestamp":
		return stat.Timestamp
	default:
		return ""
	}
}

func getReport(w http.ResponseWriter, r *http.Request) {

	var repOrder ReportRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&repOrder)

	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	request := map[string]StatsRequest{"type": StatsRequest{Timestamp: "0"}}

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	conn, err := net.Dial("tcp", IP+":"+strconv.Itoa(PORT+1))
	if err != nil {
		fmt.Println("Connecting error:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Sending error:", err)
		return
	}

	buffer := make([]byte, 32768)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	var receivedQueue Queue

	err = json.Unmarshal(buffer[:n], &receivedQueue)
	if err != nil {
		fmt.Println("Error decoding JSON-answer:", err)
		return
	}

	var existData []StatsRequest
	for !receivedQueue.isEmpty() {
		data := receivedQueue.pop()

		existData = append(existData, data)
	}

	report := buildReport(repOrder.DimensionsOrder, existData)

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write(reportJSON)
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/report", getReport)

	fmt.Println("HTTP Server listening on " + strconv.Itoa(PORT))
	err := http.ListenAndServe(IP+":"+strconv.Itoa(PORT), mux)
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
	}

}
