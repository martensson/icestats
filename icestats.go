package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/handlers"
	"github.com/mmcloughlin/geohash"
	"github.com/oschwald/geoip2-golang"
)

// XML Structs
type listener struct {
	IP        string `xml:"IP"`
	UserAgent string `xml:"UserAgent"`
	Connected string `xml:"Connected"`
	ID        string `xml:"ID"`
}

type source struct {
	XMLNAME   xml.Name   `xml:"source"`
	Mount     string     `xml:"mount,attr"`
	Listener  []listener `xml:"listener"`
	Listeners int        `xml:"Listeners"`
}

type xmlstats struct {
	XMLName xml.Name `xml:"icestats"`
	Source  source   `xml:"source"`
}

// JSON Structs
type city struct {
	Total   int
	Geohash string
}

type country struct {
	Total     int
	ISO       string
	Continent string
	Cities    map[string]*city
}

type stats struct {
	Total     int
	Countries map[string]*country
}

// Config
type config struct {
	User     string
	Password string
	URL      string
	Mounts   []string
}

var db *geoip2.Reader
var cfg config

func statHandler(w http.ResponseWriter, r *http.Request) {
	stats := new(stats)
	stats.Countries = make(map[string]*country)
	for _, mount := range cfg.Mounts {
		client := &http.Client{}
		req, err := http.NewRequest("GET", cfg.URL+"/admin/listclients?mount="+mount, nil)
		req.SetBasicAuth(cfg.User, cfg.Password)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			continue
		}
		var xmlstats xmlstats
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		xml.Unmarshal(bodyBytes, &xmlstats)
		stats.Total = stats.Total + xmlstats.Source.Listeners
		for _, l := range xmlstats.Source.Listener {
			ip := net.ParseIP(l.IP)
			record, err := db.City(ip)
			if err != nil {
				continue
			}
			co := record.Country.Names["en"]
			if co != "" {
				if _, ok := stats.Countries[co]; ok {
					stats.Countries[co].Total = stats.Countries[co].Total + 1
				} else {
					stats.Countries[co] = &country{
						ISO:       record.Country.IsoCode,
						Total:     1,
						Continent: record.Continent.Names["en"],
						Cities:    make(map[string]*city),
					}
				}
				ci := record.City.Names["en"]
				if ci != "" {
					if _, ok := stats.Countries[co].Cities[ci]; ok {
						stats.Countries[co].Cities[ci].Total = stats.Countries[co].Cities[ci].Total + 1
					} else {
						hash := geohash.Encode(record.Location.Latitude, record.Location.Longitude)
						stats.Countries[co].Cities[ci] = &city{
							Total:   1,
							Geohash: hash,
						}
					}
				}
			}
		}
	}
	b, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
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
	r.HandleFunc("/", statHandler)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	proxyRouter := handlers.ProxyHeaders(loggedRouter)
	srv := &http.Server{
		Handler:      proxyRouter,
		Addr:         *iface + ":" + *port,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	log.Println("icestats listening on", srv.Addr)
	db, err = geoip2.Open("GeoIP2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Fatal(srv.ListenAndServe())
}
