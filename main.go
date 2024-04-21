package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	// "strings"
	"time"
)

type ControlMessage struct {
	Target string
	Count  int64
}

func main() {
	controlChannel := make(chan ControlMessage)
	workerCompleteChan := make(chan bool)
	statusPollChannel := make(chan chan bool)
	workerActive := false
	go admin(controlChannel, statusPollChannel)

	for {
		select {
		case respChan := <-statusPollChannel:
			respChan <- workerActive
		case msg := <-controlChannel:
			workerActive = true
			go doStuff(msg, workerCompleteChan)
		case status := <-workerCompleteChan:
			workerActive = status
		}
	}
}

func admin(cc chan ControlMessage, statusPollChannel chan chan bool) {
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		// hostTokens := strings.Split(r.Host, ":") // ?
		r.ParseForm()
		count, err := strconv.ParseInt(r.FormValue("count"), 10, 32)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		msg := ControlMessage{
			Target: r.FormValue("target"),
			Count:  count,
		}
		cc <- msg
		fmt.Fprintf(w, "Control message issued for Target %s, count %d", html.EscapeString(r.FormValue("target")), count)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		reqChan := make(chan bool)
		statusPollChannel <- reqChan
		timeout := time.After(time.Second)

		select {
		case result := <-reqChan:
			if result {
				fmt.Fprint(w, "ACTIVE")
			} else {
				fmt.Fprint(w, "INACTIVE")
			}
			return
		case <-timeout:
			fmt.Fprint(w, "TIMEOUT")
		}
	})

	log.Fatal(http.ListenAndServe(":1337", nil))
}

// Below is original
func doStuff(msg ControlMessage, workerCompleteChan chan bool) {
	for i := int64(0); i < msg.Count; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("%s #%d\n", msg.Target, i+1)
	}
	workerCompleteChan <- false
}
