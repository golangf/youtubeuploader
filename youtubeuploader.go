package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

type chanChan chan chan struct{}

// Flags for CLI.
type Flags struct {
	Video               string
	Thumbnail           string
	Caption             string
	Meta                string
	Log                 string
	ClientID            string
	ClientToken         string
	Title               string
	Description         string
	Tags                string
	Language            string
	Category            string
	PrivacyStatus       string
	Embeddable          string
	License             string
	PublicStatsViewable string
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
	AuthHeadless        string
	Version             string
}

// Set by compile-time to match git tag
var appVersion = ""
var f = Flags{}

func parseBool(txt string, def bool) bool {
	ans, err := strconv.ParseBool(txt)
	if err != nil {
		return def
	}
	return ans
}

func parseInt(txt string, def int) int {
	ans, err := strconv.ParseInt(txt, 10, 32)
	if err != nil {
		return def
	}
	return int(ans)
}

func parseFloat(txt string, def float64) float64 {
	ans, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		return def
	}
	return ans
}

func parsePrivacyStatus(txt string) string {
	re := regexp.MustCompile("open|fee|public|common|creative")
	if txt == "" || re.FindString(txt) == "" {
		return "public"
	}
	return "private"
}

func parseLicense(txt string) string {
	re := regexp.MustCompile("open|free|common|creative")
	if txt == "" || re.FindString(txt) == "" {
		return ""
	}
	return "creativeCommon"
}

func openFile(nam string) (io.ReadCloser, int64) {
	var fil io.ReadCloser
	var siz int64
	var err error
	// open video file
	if nam != "" {
		fil, siz, err = Open(nam)
		if err != nil {
			log.Fatal(err)
		}
		defer fil.Close()
	}
	return fil, siz
}

func getEnv(f *Flags) {
	f.Log = os.Getenv("YOUTUBEUPLOADER_LOG")
	f.ClientID = os.Getenv("YOUTUBEUPLOADER_CLIENT_ID")
	f.ClientToken = os.Getenv("YOUTUBEUPLOADER_CLIENT_TOKEN")
	f.Title = os.Getenv("YOUTUBEUPLOADER_TITLE")
	f.Description = os.Getenv("YOUTUBEUPLOADER_DESCRIPTION")
	f.Tags = os.Getenv("YOUTUBEUPLOADER_TAGS")
	f.Language = os.Getenv("YOUTUBEUPLOADER_LANGUAGE")
	f.Category = os.Getenv("YOUTUBEUPLOADER_CATEGORY")
	f.PrivacyStatus = os.Getenv("YOUTUBEUPLOADER_PRIVACYSTATUS")
	f.Embeddable = os.Getenv("YOUTUBEUPLOADER_EMBEDDABLE")
	f.License = os.Getenv("YOUTUBEUPLOADER_LICENSE")
	f.PublicStatsViewable = os.Getenv("YOUTUBEUPLOADER_PUBLICSTATSVIEWABLE")
	f.PublishAt = os.Getenv("YOUTUBEUPLOADER_PUBLISHAT")
	f.RecordingDate = os.Getenv("YOUTUBEUPLOADER_RECORDINGDATE")
	f.PlaylistIds = os.Getenv("YOUTUBEUPLOADER_PLAYLISTIDS")
	f.PlaylistTitles = os.Getenv("YOUTUBEUPLOADER_PLAYLISTTITLES")
	f.LocationLatitude = os.Getenv("YOUTUBEUPLOADER_LOCATION_LATITUDE")
	f.LocationLongitude = os.Getenv("YOUTUBEUPLOADER_LOCATION_LONGITUDE")
	f.LocationDescription = os.Getenv("YOUTUBEUPLOADER_LOCATIONDESCRIPTION")
	f.UploadChunk = os.Getenv("YOUTUBEUPLOADER_UPLOAD_CHUNK")
	f.UploadRate = os.Getenv("YOUTUBEUPLOADER_UPLOAD_RATE")
	f.UploadTime = os.Getenv("YOUTUBEUPLOADER_UPLOAD_TIME")
	f.AuthPort = os.Getenv("YOUTUBEUPLOADER_AUTH_PORT")
	f.AuthHeadless = os.Getenv("YOUTUBEUPLOADER_AUTH_HEADLESS")
}

func getFlags(f *Flags) {
	flag.StringVar(&f.Video, "video", "", "set input video file")
	flag.StringVar(&f.Video, "v", "", "set input video file")
	flag.StringVar(&f.Thumbnail, "thumbnail", "", "set input thumbnail file")
	flag.StringVar(&f.Thumbnail, "t", "", "set input thumbnail file")
	flag.StringVar(&f.Caption, "caption", "", "set input caption file")
	flag.StringVar(&f.Caption, "c", "", "set input caption file")
	flag.StringVar(&f.Meta, "meta", "", "set input meta file")
	flag.StringVar(&f.Meta, "m", "", "set input meta file")
	flag.StringVar(&f.Log, "log", "", "enable log")
	flag.StringVar(&f.Log, "l", "", "enable log")
	flag.StringVar(&f.ClientID, "client_id", "", "set client id credentials path")
	flag.StringVar(&f.ClientID, "ci", "", "set client id credentials path")
	flag.StringVar(&f.ClientToken, "client_token", "", "set client token credentials path")
	flag.StringVar(&f.ClientToken, "ct", "", "set client token credentials path")
	flag.StringVar(&f.Title, "video_title", "", "set video title")
	flag.StringVar(&f.Title, "vt", "", "set video title")
	flag.StringVar(&f.Description, "video_description", "", "set video description")
	flag.StringVar(&f.Description, "vd", "", "set video description")
	flag.StringVar(&f.Tags, "video_tags", "", "set video tags/keywords")
	flag.StringVar(&f.Tags, "vk", "", "set video tags/keywords")
	flag.StringVar(&f.Language, "video_language", "en", "set video language")
	flag.StringVar(&f.Language, "vl", "en", "set video language")
	flag.StringVar(&f.Category, "video_category", "", "set video category id")
	flag.StringVar(&f.Category, "vc", "", "set video category id")
	flag.StringVar(&f.PrivacyStatus, "video_privacystatus", "public", "set video privacy status")
	flag.StringVar(&f.PrivacyStatus, "vp", "public", "set video privacy status")
	flag.StringVar(&f.Embeddable, "video_embeddable", "", "enable video to be embeddable")
	flag.StringVar(&f.Embeddable, "ve", "", "enable video to be embeddable")
	flag.StringVar(&f.License, "video_license", "standard", "set video license")
	flag.StringVar(&f.License, "vl", "standard", "set video license")
	flag.StringVar(&f.PublicStatsViewable, "video_publicstatsviewable", "", "enable public video stats to be viewable")
	flag.StringVar(&f.PublicStatsViewable, "vs", "", "enable public video stats to be viewable")
	flag.StringVar(&f.PublishAt, "video_publishat", "", "set video publish time")
	flag.StringVar(&f.PublishAt, "vpa", "", "set video publish time")
	flag.StringVar(&f.RecordingDate, "video_recordingdate", "", "set video recording date")
	flag.StringVar(&f.RecordingDate, "vrd", "", "set video recording date")
	flag.StringVar(&f.PlaylistIds, "video_playlistids", "", "set video playlist ids")
	flag.StringVar(&f.PlaylistIds, "vpi", "", "set video playlist ids")
	flag.StringVar(&f.PlaylistTitles, "video_playlisttitles", "", "set video playlist titles")
	flag.StringVar(&f.PlaylistTitles, "vpt", "", "set video playlist titles")
	flag.StringVar(&f.LocationLatitude, "video_location_latitude", "", "set video latitude coordinate")
	flag.StringVar(&f.LocationLatitude, "vla", "", "set video latitude coordinate")
	flag.StringVar(&f.LocationLongitude, "video_location_longitude", "", "set video longitude coordinate")
	flag.StringVar(&f.LocationLongitude, "vlo", "", "set video longitude coordinate")
	flag.StringVar(&f.LocationDescription, "video_locationdescription", "", "set video location description")
	flag.StringVar(&f.LocationDescription, "vld", "", "set video location description")
	flag.StringVar(&f.UploadChunk, "upload_chunk", "", "set upload chunk size in bytes")
	flag.StringVar(&f.UploadChunk, "uc", "", "set upload chunk size in bytes")
	flag.StringVar(&f.UploadRate, "upload_rate", "", "set upload rate limit in kbps")
	flag.StringVar(&f.UploadRate, "ur", "", "set upload rate limit in kbps")
	flag.StringVar(&f.UploadTime, "upload_time", "", "set upload time limit ex- \"10:00-14:00\"")
	flag.StringVar(&f.UploadTime, "ut", "", "set upload time limit ex- \"10:00-14:00\"")
	flag.StringVar(&f.AuthPort, "auth_port", "", "set OAuth request port")
	flag.StringVar(&f.AuthPort, "ap", "", "set OAuth request port")
	flag.StringVar(&f.AuthHeadless, "auth_headless", "", "enable browserless OAuth process")
	flag.StringVar(&f.AuthHeadless, "ah", "", "enable browserless OAuth process")
	flag.StringVar(&f.Version, "version", "", "show version")
	flag.Parse()
}

func getUploadTime(f *Flags) limitRange {
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

func onVersion(f *Flags) {
	if parseBool(f.Version, false) {
		fmt.Printf("youtubeuploader v%s\n", appVersion)
		os.Exit(0)
	}
}

func onHelp(f *Flags) {
	if f.Video == "" && f.Title == "" {
		fmt.Printf("No video file to upload!\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func onTitle(srv *youtube.Service, txt string) {
	res, err := srv.Search.List("snippet").Type("video").Q(txt).Do()
	if err != nil {
		if res != nil {
			log.Fatalf("Error searching video title  '%v': %v, %v", txt, err, res.HTTPStatusCode)
		} else {
			log.Fatalf("Error searching video title '%v': %v", txt, err)
		}
	}
	var re = regexp.MustCompile("\\W")
	var texta = strings.ToLower(re.ReplaceAllString(txt, ""))
	for _, item := range res.Items {
		var textb = strings.ToLower(re.ReplaceAllString(item.Snippet.Title, ""))
		if texta == textb {
			fmt.Printf("%v\n", item.Id.VideoId)
		}
	}
}

func updateFlags(f *Flags) {
	if f.Title == "" {
		f.Title = f.Video
	}
	if f.Description == "" {
		f.Description = f.Video
	}
}

func updateMeta(m *youtube.Video, f *Flags) {
	if m.Snippet.Title == "" {
		m.Snippet.Title = f.Title
	}
	if m.Snippet.Description == "" {
		m.Snippet.Description = f.Description
	}
	if m.Snippet.Tags == nil && strings.Trim(f.Tags, "") != "" {
		m.Snippet.Tags = strings.Split(f.Tags, ",")
	}
	if m.Snippet.DefaultLanguage == "" && f.Language != "" {
		m.Snippet.DefaultLanguage = f.Language
	}
	if m.Snippet.DefaultAudioLanguage == "" && f.Language != "" {
		m.Snippet.DefaultAudioLanguage = f.Language
	}
	if m.Snippet.CategoryId == "" && f.Category != "" {
		m.Snippet.CategoryId = f.Category
	}
	if m.Status.PrivacyStatus == "" {
		m.Status.PrivacyStatus = f.PrivacyStatus
	}
	if f.License != "" {
		m.Status.License = parseLicense(f.License)
	}
	if f.PublicStatsViewable != "" {
		m.Status.PublicStatsViewable = parseBool(f.PublicStatsViewable, false)
	}
	if f.PublishAt != "" {
		m.Status.PublishAt = f.PublishAt
	}
	if f.RecordingDate != "" {
		m.RecordingDetails.RecordingDate = f.RecordingDate
	}
	if f.LocationLatitude != "" {
		m.RecordingDetails.Location.Latitude = parseFloat(f.LocationLatitude, math.NaN())
	}
	if f.LocationLongitude != "" {
		m.RecordingDetails.Location.Longitude = parseFloat(f.LocationLongitude, math.NaN())
	}
	if f.LocationDescription != "" {
		m.RecordingDetails.LocationDescription = f.LocationDescription
	}
}

func uploadVideo(srv *youtube.Service, nam string, fil io.ReadCloser, obj *youtube.Video, cnk int, cquit chanChan) *youtube.Video {
	fmt.Printf("Uploading file '%s'...\n", nam)
	opt := googleapi.ChunkSize(cnk)
	req := srv.Videos.Insert("snippet,status,recordingDetails", obj)
	res, err := req.Media(fil, opt).Do()
	if cquit != nil {
		quit := make(chan struct{})
		cquit <- quit
		<-quit
	}
	if err != nil {
		if res != nil {
			log.Fatalf("Error making YouTube API call: %v, %v", err, res.HTTPStatusCode)
		} else {
			log.Fatalf("Error making YouTube API call: %v", err)
		}
	}
	fmt.Printf("Upload successful! Video ID: %v\n", res.Id)
	return res
}

func uploadThumbnail(srv *youtube.Service, id string, nam string, fil io.ReadCloser) {
	if fil != nil {
		log.Printf("Uploading thumbnail '%s'...\n", nam)
		res, err := srv.Thumbnails.Set(id).Media(fil).Do()
		if err != nil {
			if res != nil {
				log.Fatalf("Error uploading thumbnail: %v, %v", err, res.HTTPStatusCode)
			} else {
				log.Fatalf("Error uploading thumbnail: %v", err)
			}
		}
		fmt.Printf("Thumbnail uploaded!\n")
	}
}

func uploadCaption(srv *youtube.Service, id string, nam string, fil io.ReadCloser) {
	if fil != nil {
		c := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{},
		}
		c.Snippet.VideoId = id
		c.Snippet.Language = f.Language
		c.Snippet.Name = f.Language
		log.Printf("Uploading caption '%s'...\n", nam)
		req := srv.Captions.Insert("snippet", c).Sync(true)
		res, err := req.Media(fil).Do()
		if err != nil {
			if res != nil {
				log.Fatalf("Error uploading caption: %v, %v", err, res.HTTPStatusCode)
			} else {
				log.Fatalf("Error uploading caption: %v", err)
			}
		}
		fmt.Printf("Caption uploaded!\n")
	}
}

func addToPlaylistID(srv *youtube.Service, pid string, sta string, id string) {
	p := Playlistx{}
	if sta != "" {
		p.PrivacyStatus = sta
	}
	// PlaylistID is deprecated in favour of PlaylistIDs
	if pid != "" {
		p.Id = pid
		err := p.AddVideoToPlaylist(srv, id)
		if err != nil {
			log.Fatalf("Error adding video to playlist: %s", err)
		}
	}
}

func addToPlaylistIDs(srv *youtube.Service, pids []string, sta string, id string) {
	p := Playlistx{}
	if sta != "" {
		p.PrivacyStatus = sta
	}
	if len(pids) > 0 {
		p.Title = ""
		for _, pid := range pids {
			p.Id = pid
			err := p.AddVideoToPlaylist(srv, id)
			if err != nil {
				log.Fatalf("Error adding video to playlist: %s", err)
			}
		}
	}
}

func addToPlaylistTitles(srv *youtube.Service, pnams []string, sta string, id string) {
	p := Playlistx{}
	if sta != "" {
		p.PrivacyStatus = sta
	}
	if len(pnams) > 0 {
		p.Id = ""
		for _, nam := range pnams {
			p.Title = nam
			err := p.AddVideoToPlaylist(srv, id)
			if err != nil {
				log.Fatalf("Error adding video to playlist: %s", err)
			}
		}
	}
}

// Main.
func main() {
	getEnv(&f)
	getFlags(&f)
	onVersion(&f)
	onHelp(&f)
	uploadTime := getUploadTime(&f)
	videoFile, fileSize := openFile(f.Video)
	thumbnailFile, _ := openFile(f.Thumbnail)
	captionFile, _ := openFile(f.Caption)

	ctx := context.Background()
	transport := &limitTransport{rt: http.DefaultTransport, lr: uploadTime, filesize: fileSize}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: transport,
	})

	var quitChan chanChan
	if parseBool(f.Log, false) {
		quitChan = make(chanChan)
		go func() {
			Progress(quitChan, transport, fileSize)
		}()
	}
	client, err := buildOAuthHTTPClient(ctx, []string{youtube.YoutubeUploadScope, youtube.YoutubepartnerScope, youtube.YoutubeScope})
	if err != nil {
		log.Fatalf("Error building OAuth client: %v", err)
	}

	upload := &youtube.Video{
		Snippet:          &youtube.VideoSnippet{},
		RecordingDetails: &youtube.VideoRecordingDetails{},
		Status:           &youtube.VideoStatus{},
	}

	videoMeta := LoadVideoMeta(f.Meta, upload)

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating YouTube client: %s", err)
	}
	// search video by title
	if f.Video == "" && f.Title != "" {
		onTitle(service, f.Title)
		os.Exit(0)
	}
	updateFlags(&f)
	updateMeta(upload, &f)
	video := uploadVideo(service, f.Video, videoFile, upload, parseInt(f.UploadChunk, 0), quitChan)
	uploadThumbnail(service, video.Id, f.Thumbnail, thumbnailFile)
	uploadCaption(service, video.Id, f.Caption, captionFile)
	addToPlaylistID(service, videoMeta.PlaylistID, upload.Status.PrivacyStatus, video.Id)
	addToPlaylistIDs(service, videoMeta.PlaylistIDs, upload.Status.PrivacyStatus, video.Id)
	addToPlaylistTitles(service, videoMeta.PlaylistTitles, upload.Status.PrivacyStatus, video.Id)
}
