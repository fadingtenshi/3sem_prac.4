package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const HASH_MAX_SIZE = 50

var links HashTable = *NewHashTable()

const (
	PORT = 6379
)

type S = string

type HashTab struct {
	key  S
	data S
	used bool
}

type HashTable struct {
	hTab []*HashTab
}

func NewHashTable() *HashTable {
	hTab := make([]*HashTab, HASH_MAX_SIZE)
	for i := range hTab {
		hTab[i] = &HashTab{}
	}
	return &HashTable{hTab: hTab}
}

func simpleHash(str string) int {
	hash := 0
	if len(str) == 0 {
		return -1
	}
	for _, element := range str {
		hash += int(element)
	}
	return hash
}

func insert(container *HashTable, new_key, new_data S, w http.ResponseWriter) string {
	new_key_ := simpleHash(new_key)
	hash := new_key_ % HASH_MAX_SIZE

	var outMessage string

	if new_key_ == -1 {

		return "ZLK"

	} else if container.hTab[hash].key == new_key {
		return container.hTab[hash].key

	} else {

		initialHash := hash
		for {
			if !container.hTab[hash].used {
				container.hTab[hash].data = new_data
				container.hTab[hash].key = new_key
				container.hTab[hash].used = true
				outMessage = "Data was inserted"
				fmt.Println(outMessage)
				return "OK"
			}
			hash = (hash + 1) % HASH_MAX_SIZE
			if hash == initialHash {
				break
			}
		}
		return "NFS"
	}
}

func get(container *HashTable, key S) S {
	key_ := simpleHash(key)
	var outMessage string

	if key_ == -1 {
		fmt.Println("Zero-length key")
		return ""
	}

	hash := key_ % HASH_MAX_SIZE
	initialHash := hash
	for {
		if container.hTab[hash].key == key && container.hTab[hash].used {
			return container.hTab[hash].data
		}
		hash = (hash + 1) % HASH_MAX_SIZE
		if hash == initialHash {
			break
		}
	}
	outMessage = "Key not found"
	fmt.Println(outMessage)
	return ""
}

type StatsRequest struct {
	ID          int    `json:"recordId"`
	PID         int    `json:"parent_id"`
	RedirectURL string `json:"redirect_url"`
	IPAddress   string `json:"ip_address"`
	Timestamp   string `json:"timestamp"`
	Count       int    `json:"redirect_count"`
}

func shortenLink(link string) string {

	hasher := sha256.New()
	hasher.Write([]byte(link))
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash[:7]

}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form",
				http.StatusInternalServerError)
			return
		}

		link := r.Form.Get("link")

		fmt.Println("client POST query body: ", link)

		resp, err := http.Head(link)
		if err != nil {

			fmt.Println("Error:", err)
			fmt.Fprintf(w, "Invalid URL")
			return

		}
		defer resp.Body.Close()

		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {

			fmt.Fprintf(w, "Invalid URL")
			return

		}

		shortLink := shortenLink(link)

		ans := insert(&links, shortLink, link, w)

		if ans == "ZLK" {

			fmt.Fprintf(w, "Zero-length key")

		} else if ans == "OK" {

			fmt.Fprintf(w, "http://localhost:6379/"+shortLink)

		} else if ans == "NFS" {

			fmt.Fprintf(w, "There's no free slot in the HashTable")

		} else {

			fmt.Fprintf(w, "This link has already been shortened: "+"http://localhost:6379/"+ans)

		}
	}

	if r.Method == "GET" {
		path := strings.TrimPrefix(r.URL.Path, "/")
		fmt.Println("client GET query body: " + path)

		usualLink := get(&links, path)

		var stat StatsRequest

		if usualLink != "" {

			currTime := time.Now().Format("15:04")

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				fmt.Println("Error extracting IP address:", err)
				return
			}

			// Преобразование IPv6 в IPv4
			ipAddr := net.ParseIP(ip)
			ipv4 := ipAddr.To4()
			if ipv4 != nil {
				ip = ipv4.String()
			}

			urlAns := usualLink + "(" + path + ")"

			a := 1
			b := 4

			stat.IPAddress = strconv.Itoa(rand.Intn(b-a) + a)

			stat.RedirectURL = urlAns
			stat.Timestamp = currTime
			jsonData, err := json.Marshal(stat)
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				return
			}
			response, err := http.Post("http://localhost:"+strconv.Itoa(PORT+1)+"/", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Error sending POST request:", err)
				return
			}
			defer response.Body.Close()
			http.Redirect(w, r, usualLink, http.StatusFound)

		} else {
			http.NotFound(w, r)
		}
	}
}

func main() {

	fmt.Println("Listening on " + strconv.Itoa(PORT) + "...")

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	http.ListenAndServe(":"+strconv.Itoa(PORT), mux)

}
