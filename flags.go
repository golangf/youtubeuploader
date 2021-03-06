package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

//
// Types
//
type appFlags struct {
	Help                bool
	Version             bool
	Log                 bool
	Id                  string
	Video               string
	Thumbnail           string
	Caption             string
	DescriptionPath     string
	Meta                string
	ClientID            string
	ClientToken         string
	Title               string
	Description         string
	Tags                string
	Language            string
	Category            string
	PrivacyStatus       string
	Embeddable          bool
	License             string
	PublicStatsViewable bool
	PublishAt           string
	RecordingDate       string
	PlaylistIds         string
	PlaylistTitles      string
	LocationLatitude    string
	LocationLongitude   string
	LocationDescription string
	UploadChunk         string
	UploadRate          string
	UploadTime          string
	AuthPort            string
	AuthHeadless        bool
}
type boolFlag struct {
	Short string
	Usage string
	Value *bool
}
type stringFlag struct {
	Short string
	Usage string
	Value *string
}

//
// Global variables
//
var f = appFlags{}
var fBool = map[string]boolFlag{
	"log":                 {"l", "enable log", &f.Log},
	"embeddable":          {"oe", "enable video to be embeddable", &f.Embeddable},
	"publicstatsviewable": {"os", "enable public video stats to be viewable", &f.PublicStatsViewable},
	"auth_headless":       {"ah", "enable browserless OAuth process", &f.AuthHeadless},
}
var fString = map[string]stringFlag{
	"id":                  {"i", "set video id", &f.Id},
	"video":               {"v", "set input video file", &f.Video},
	"thumbnail":           {"t", "set input thumbnail file", &f.Thumbnail},
	"caption":             {"c", "set input caption file", &f.Caption},
	"descriptionpath":     {"d", "set input description file", &f.DescriptionPath},
	"meta":                {"m", "set input meta file", &f.Meta},
	"client_id":           {"ci", "set client id credentials path (client_id.json)", &f.ClientID},
	"client_token":        {"ct", "set client token credentials path (client_token.json)", &f.ClientToken},
	"title":               {"ot", "set video title (video)", &f.Title},
	"description":         {"od", "set video description (video)", &f.Description},
	"tags":                {"ok", "set video tags/keywords", &f.Tags},
	"language":            {"ol", "set video language (en)", &f.Language},
	"category":            {"oc", "set video category (people and blogs)", &f.Category},
	"privacystatus":       {"op", "set video privacy status (private)", &f.PrivacyStatus},
	"license":             {"oi", "set video license (standard)", &f.License},
	"publishat":           {"opa", "set video publish time", &f.PublishAt},
	"recordingdate":       {"ord", "set video recording date", &f.RecordingDate},
	"playlistids":         {"opi", "set video playlist ids", &f.PlaylistIds},
	"playlisttitles":      {"opt", "set video playlist titles", &f.PlaylistTitles},
	"location_latitude":   {"ola", "set video latitude coordinate", &f.LocationLatitude},
	"location_longitude":  {"olo", "set video longitude coordinate", &f.LocationLongitude},
	"locationdescription": {"old", "set video location description", &f.LocationDescription},
	"upload_chunk":        {"uc", "set upload chunk size in bytes", &f.UploadChunk},
	"upload_rate":         {"ur", "set upload rate limit in kbps", &f.UploadRate},
	"upload_time":         {"ut", "set upload time limit ex- \"10:00-14:00\"", &f.UploadTime},
	"auth_port":           {"ap", "set OAuth request port (8080)", &f.AuthPort},
}

//
// Functions
//
func getFlagsDynamic() {
	var ai = strings.Split(f.ClientID, ";")
	var at = strings.Split(f.ClientToken, ";")
	var n = rand.Intn(65535)
	f.ClientID = ai[n%len(ai)]
	f.ClientToken = at[n%len(at)]
	if f.Description == "" && f.DescriptionPath != "" {
		dat, err := ioutil.ReadFile(f.DescriptionPath)
		if err != nil {
			log.Fatalf("Error reading description file '%v': %v", f.DescriptionPath, err)
		}
		f.Description = string(dat)
	}
}

func getFlagsBasic() {
	f.ClientID = parseString(f.ClientID, "client_id.json")
	f.ClientToken = parseString(f.ClientToken, "client_token.json")
	f.AuthPort = parseString(f.AuthPort, "8080")
}

func getFlags() {
	rand.Seed(time.Now().Unix())
	flag.BoolVar(&f.Help, "help", false, "show help")
	flag.BoolVar(&f.Version, "version", false, "show version")
	for k, bf := range fBool {
		var bv = parseBool(os.Getenv("YOUTUBEUPLOADER_"+strings.ToUpper(k)), false)
		flag.BoolVar(bf.Value, bf.Short, bv, bf.Usage)
		flag.BoolVar(bf.Value, k, bv, bf.Usage)
	}
	for k, sf := range fString {
		flag.StringVar(sf.Value, sf.Short, "", sf.Usage)
		flag.StringVar(sf.Value, k, "", sf.Usage)
	}
	for k, sf := range fString {
		*sf.Value = parseString(*sf.Value, os.Getenv("YOUTUBEUPLOADER_"+strings.ToUpper(k)))
	}
	flag.Parse()
	getFlagsBasic()
	getFlagsDynamic()
}

func getUploadTime() limitRange {
	var ans limitRange
	var err error
	if f.UploadTime != "" {
		ans, err = parseLimitBetween(f.UploadTime)
		if err != nil {
			fmt.Printf("Invalid upload time: %v", err)
			os.Exit(1)
		}
	}
	return ans
}
