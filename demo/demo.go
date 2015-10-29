package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wannaup/paparazzogo"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("MJPEG restreamer forked from https://github.com/putsi/paparazzogo")
	fmt.Println("Wannaup srls - 2015")
	// Local server settings
	path2d := "/2d.jpg"
	path3d := "/3d.jpg"
	addr := ""
	streamUrls := map[string]string{}
	// MJPEG-stream settings
	user := ""
	pass := ""

	//get server to ask endpoint urls
	var whoKnows = flag.String("configserver", "http://127.0.0.1:8000", "The address of the server who knows the config")
	var timeout = flag.Int("grace", 30, "Timeout(s) after which* the stream will be closed if no requests")
	var lPort = flag.Int("port", 5050, "Port to listen to (also edit config.json for WUI)")
	flag.Parse()
	//set listening port
	addr = fmt.Sprintf(":%d", *lPort)
	fmt.Print("listening on " + addr)
	//get endpoints urls
	gotit := false
	for !gotit {
		time.Sleep(3 * time.Second)
		resp, err := http.Get(*whoKnows + "/api/settings")
		if err != nil {
			fmt.Println("Could not contact config server, retrying...")
			continue
		}
		defer resp.Body.Close()
		//parse response and set config
		body, errr := ioutil.ReadAll(resp.Body)
		if errr != nil {
			fmt.Println("Error getting resp body")
			continue
		}
		err = json.Unmarshal(body, &streamUrls)
		if err != nil {
			fmt.Println("Error parsing resp body")
			fmt.Println(err)
			continue
		}
		gotit = true
	}
	fmt.Println(streamUrls)
	fmt.Println("Got stream urls, starting ")
	// If there is zero GET-requests for 30 seconds, mjpeg-stream will be closed.
	// Streaming will be reopened after next request.
	tout := time.Duration(*timeout) * time.Second
	mjpegStream2 := streamUrls["videoSource"]
	mjpegHandler2 := paparazzogo.NewMjpegproxy()
	mjpegHandler2.OpenStream(mjpegStream2, user, pass, tout)
	mjpegStream3 := streamUrls["depthSource"]
	mjpegHandler3 := paparazzogo.NewMjpegproxy()
	mjpegHandler3.OpenStream(mjpegStream3, user, pass, tout)

	mux := http.NewServeMux()
	mux.Handle(path2d, mjpegHandler2)
	mux.Handle(path3d, mjpegHandler3)

	s := &http.Server{
		Addr:    addr,
		Handler: mux,
		// Read- & Write-timeout prevent server from getting overwhelmed in idle connections
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(s.ListenAndServe())

	block := make(chan bool)
	// time.Sleep(time.Second * 30)
	// mp.CloseStream()
	// mp.OpenStream(newMjpegstream, newUser, newPass, newTimeout)
	<-block

}
