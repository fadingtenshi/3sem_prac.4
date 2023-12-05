package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
)

var queue = &Queue{}
var wg sync.WaitGroup

type S string

type StatsRequest struct {
	ID          int    `json:"recordId"`
	PID         int    `json:"parent_id"`
	RedirectURL string `json:"redirect_url"`
	IPAddress   string `json:"ip_address"`
	Timestamp   string `json:"timestamp"`
	Count       int    `json:"redirect_count"`
}
type Node struct {
	Data StatsRequest `json:"data"`
	Next *Node        `json:"next"`
}

type Queue struct {
	First *Node `json:"first"`
	Last  *Node `json:"last"`
}

func (q *Queue) push(newData StatsRequest) {

	newNode := &Node{Data: newData, Next: nil}

	if q.isEmpty() {
		q.First = newNode
	} else {
		q.Last.Next = newNode
	}

	q.Last = newNode
}

func (q *Queue) isEmpty() bool {
	return q.First == nil
}

const (
	PORT = 6381
	IP   = "192.168.177.235"
)

func handlerFunc(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 32768)

	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}

	var request map[string]StatsRequest
	err = json.Unmarshal(buffer[:n], &request)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	if request["type"].Timestamp != "0" {

		queue.push(request["type"])

	} else {

		queueJSON, err := json.MarshalIndent(queue, "", "    ")
		if err != nil {
			fmt.Println("Error encoding JSON response:", err)
			return
		}

		_, err = conn.Write(queueJSON)
		if err != nil {
			fmt.Println("Error sending data:", err)
			return
		}
	}
}

func main() {

	listen, err := net.Listen("tcp", IP+":"+strconv.Itoa(PORT))
	if err != nil {
		fmt.Println("Error listening", err)
		return
	}
	defer listen.Close()

	wg.Add(1)

	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}

			wg.Add(1)

			go handlerFunc(conn)
		}
	}()

	wg.Wait()

}
