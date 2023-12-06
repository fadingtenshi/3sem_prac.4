package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	PORT = 6379
	IP   = "192.168.177.235" // Your IPv4
)

type ReportRequest struct {
	DimensionsOrder []string `json:"dimensionsOrder"`
}

func main() {

	fmt.Println("Client started ...")
	fmt.Println("Dimensions: [RedirectURL IPAddress Timestamp]")
	fmt.Println("Use /short/[...] to short your link")
	fmt.Println("Enter /get/[shorted link] to get the original link")
	fmt.Println("Enter /report/[dimensions] to get report")

	for {

		var text string
		fmt.Println("Enter command: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text = scanner.Text()

		if !(strings.HasPrefix(text, "/short/")) && !(strings.HasPrefix(text, "/get")) &&
			!(strings.HasPrefix(text, "/report/")) {
			fmt.Println("Invalid method")
			continue
		}

		method := strings.SplitN(text, "/", 3)

		if method[1] == "short" {

			resp, err := http.PostForm("http://"+IP+":"+strconv.Itoa(PORT)+"/", url.Values{"link": {method[2]}})

			if err != nil {
				fmt.Println("Error sending HTTP request:", err)
				continue
			}

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading HTTP response:", err)
				continue
			}

			fmt.Println("Message from server:", string(body))

		} else if method[1] == "get" {

			resp, err := http.Get(method[2])

			if err != nil {
				fmt.Println("Error sending HTTP request:", err)
				continue
			}

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading HTTP response:", err)
				continue
			}

			fmt.Println("Message from server:", string(body))

		} else if method[1] == "report" {

			dimensions := strings.Split(method[2], " ")

			repOrder := ReportRequest{
				DimensionsOrder: dimensions,
			}

			jsonData, err := json.Marshal(repOrder)
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				return
			}
			response, err := http.Post("http://"+IP+":"+strconv.Itoa(PORT+1)+"/report", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Error sending POST request:", err)
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading HTTP response:", err)
				return
			}

			err = os.WriteFile("report.json", body, 0644)
			if err != nil {
				fmt.Println("Error writing JSON to file:", err)
				return
			}

		}

	}
}
