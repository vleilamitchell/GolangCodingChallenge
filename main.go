package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type Request struct {
	Request string
	Whitelist []string
}

type Response struct {
	Authorized bool
}

func main() {
	UpdateGeoIP()
	HandleRequests()
}

//updates the locally stored GeoLite2-Country db
func UpdateGeoIP() {
	os.Chdir("./data")
	cmd := exec.Command("geoipupdate", "1")
	err := cmd.Run()
	if err != nil {
		if fileExists("GeoLite2-Country.mmdb"){
			log.Println(err.Error())
			log.Println("Using cached file.")
		} else {
			log.Fatal(err.Error())
		}
	}
}

//sets up router and assigns handlers
func HandleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/health", healthHandler)
	router.HandleFunc("/readiness", readinessHandler)
	router.HandleFunc("/processIpRequest", processIpRequest).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         ":10000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("Starting endpoint")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	waitForShutdown(srv)
}

//handler for the ip authorization request
//request is submitted in json POST body
func processIpRequest(w http.ResponseWriter, r *http.Request){
	reqBody, _ := ioutil.ReadAll(r.Body)
	var request Request
	json.Unmarshal(reqBody, &request)
	log.Println("Request Recieved for IP: " + request.Request)
	var country = getIPCountryName(net.ParseIP(request.Request))
	var response Response
	response.Authorized = contains(request.Whitelist, country)
	json.NewEncoder(w).Encode(response)
}

//processes an IP against the geoip2 db
//returns a string with the plaintext name of the Country the IP is associated with
func getIPCountryName(ip net.IP) string{
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	record, err := db.Country(ip)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("English country name: %v\n", record.Country.Names["en"])
	return record.Country.Names["en"]
}

//utility function for finding a string in an array
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

//checks for file existence
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//health handler for Kubernetes
func healthHandler (w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
}

//readiness handler for Kubernetes
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
