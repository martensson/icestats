package main

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/handlers"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Icecast XML Structs
type listener struct {
	IP        string `xml:"IP"`
	UserAgent string `xml:"UserAgent"`
	Connected string `xml:"Connected"`
	ID        string `xml:"ID"`
}

type source struct {
	XMLNAME     xml.Name   `xml:"source"`
	Mount       string     `xml:"mount,attr"`
	Listener    []listener `xml:"listener"`
	Listeners   int        `xml:"listeners"`
	Connected   int        `xml:"Connected"`
	ContentType string     `xml:"content-type"`
}

type icestats struct {
	XMLName xml.Name `xml:"icestats"`
	Source  []source `xml:"source"`
}

// Prometheus Struct
type clients struct {
	City      string
	Country   string
	ISO       string
	Continent string
	Geohash   string
	Mount     string
	Total     float64
}

// Config
type config struct {
	User     string
	Password string
	URL      string
	Geoip2   string
}

var db *geoip2.Reader
var cfg config

func getStats(uri string) (icestats, error) {
	var icestats icestats
	client := &http.Client{}
	req, err := http.NewRequest("GET", cfg.URL+uri, nil)
	req.SetBasicAuth(cfg.User, cfg.Password)
	resp, err := client.Do(req)
	if err != nil {
		return icestats, err
	}
	if resp.StatusCode != 200 {
		return icestats, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return icestats, err
	}
	err = xml.Unmarshal(bodyBytes, &icestats)
	if err != nil {
		return icestats, err
	}
	return icestats, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Icestats Prometheus Exporter\n"))
}

func main() {
	port := flag.String("p", "8080", "Listening port")
	iface := flag.String("i", "0.0.0.0", "Listening interface")
	configtoml := flag.String("f", "icestats.toml", "Path to config. (default icestats.toml)")
	flag.Parse()
	file, err := ioutil.ReadFile(*configtoml)
	if err != nil {
		log.Fatal(err)
	}
	err = toml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	r := http.NewServeMux()
	r.HandleFunc("/", rootHandler)
	r.Handle("/metrics", promhttp.Handler())
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	proxyRouter := handlers.ProxyHeaders(loggedRouter)
	srv := &http.Server{
		Handler:      proxyRouter,
		Addr:         *iface + ":" + *port,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	log.Println("icestats listening on", srv.Addr)
	db, err = geoip2.Open(cfg.Geoip2)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	icestats := newIcestatsCollector()
	prometheus.MustRegister(icestats)
	log.Fatal(srv.ListenAndServe())
}
