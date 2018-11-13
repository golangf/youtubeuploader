package main

import "google.golang.org/api/youtube/v3"
import "google.golang.org/api/googleapi"
import "golang.org/x/oauth2"
import "net/http"
import "strings"
import "strconv"
import "context"
import "regexp"
import "flag"
import "fmt"
import "log"
import "io"
import "os"

type chanChan chan chan struct{}

// Flags for CLI.
type Flags struct {
	Video               string
	Thumbnail           string
	Caption             string
	Meta                string
	Log                 bool
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
	LocationLatitude    float64
	LocationLongitude   float64
	LocationDescription string
	UploadChunk         int
	UploadRate          float64
	UploadTime          string
	AuthPort            int
	AuthHeadless        bool
	Version             bool
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
	f.Log = parseBool(os.Getenv("YOUTUBEUPLOADER_LOG"), false)
	f.ClientID = os.Getenv("YOUTUBEUPLOADER_CLIENT_ID")
	f.ClientToken = os.Getenv("YOUTUBEUPLOADER_CLIENT_TOKEN")
	f.Title = os.Getenv("YOUTUBEUPLOADER_VIDEO_TITLE")
	f.Description = os.Getenv("YOUTUBEUPLOADER_VIDEO_DESCRIPTION")
	f.Tags = os.Getenv("YOUTUBEUPLOADER_VIDEO_TAGS")
	f.Language = os.Getenv("YOUTUBEUPLOADER_VIDEO_LANGUAGE")
	f.Category = os.Getenv("YOUTUBEUPLOADER_VIDEO_CATEGORY")
	f.PrivacyStatus = os.Getenv("YOUTUBEUPLOADER_VIDEO_PRIVACYSTATUS")
	f.Embeddable = parseBool(os.Getenv("YOUTUBEUPLOADER_VIDEO_EMBEDDABLE"), true)
	f.License = os.Getenv("YOUTUBEUPLOADER_VIDEO_LICENSE")
	f.PublicStatsViewable = parseBool(os.Getenv("YOUTUBEUPLOADER_VIDEO_PUBLICSTATSVIEWABLE"), true)
	f.PublishAt = os.Getenv("YOUTUBEUPLOADER_VIDEO_PUBLISHAT")
	f.RecordingDate = os.Getenv("YOUTUBEUPLOADER_VIDEO_RECORDINGDATE")
	f.PlaylistIds = os.Getenv("YOUTUBEUPLOADER_VIDEO_PLAYLISTIDS")
	f.PlaylistTitles = os.Getenv("YOUTUBEUPLOADER_VIDEO_PLAYLISTTITLES")
	f.LocationLatitude = parseFloat(os.Getenv("YOUTUBEUPLOADER_VIDEO_LOCATION_LATITUDE"), 0)
	f.LocationLongitude = parseFloat(os.Getenv("YOUTUBEUPLOADER_VIDEO_LOCATION_LONGITUDE"), 0)
	f.LocationDescription = os.Getenv("YOUTUBEUPLOADER_VIDEO_LOCATIONDESCRIPTION")
	f.UploadChunk = parseInt(os.Getenv("YOUTUBEUPLOADER_UPLOAD_CHUNK"), 8388608)
	f.UploadRate = parseFloat(os.Getenv("YOUTUBEUPLOADER_UPLOAD_RATE"), 0)
	f.UploadTime = os.Getenv("YOUTUBEUPLOADER_UPLOAD_TIME")
	f.AuthPort = parseInt(os.Getenv("YOUTUBEUPLOADER_AUTH_PORT"), 8080)
	f.AuthHeadless = parseBool(os.Getenv("YOUTUBEUPLOADER_AUTH_HEADLESS"), false)
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
	flag.BoolVar(&f.Log, "log", false, "enable log")
	flag.BoolVar(&f.Log, "l", false, "enable log")
	flag.StringVar(&f.ClientID, "client_id", "client_id.json", "set client id credentials path")
	flag.StringVar(&f.ClientID, "ci", "client_id.json", "set client id credentials path")
	flag.StringVar(&f.ClientToken, "client_token", "client_token.json", "set client token credentials path")
	flag.StringVar(&f.ClientToken, "ct", "client_token.json", "set client token credentials path")
	flag.StringVar(&f.Title, "video_title", "", "set video title")
	flag.StringVar(&f.Title, "vt", "", "set video title")
	flag.StringVar(&f.Description, "video_description", "", "set video description")
	flag.StringVar(&f.Description, "vd", "", "set video description")
	flag.StringVar(&f.Tags, "video_tags", "", "set video tags/keywords")
	flag.StringVar(&f.Tags, "vk", "", "set video tags/keywords")
	flag.StringVar(&f.Language, "video_language", "en", "set video language")
	flag.StringVar(&f.Language, "vl", "en", "set video language")
	flag.StringVar(&f.Category, "video_category", "22", "set video category id")
	flag.StringVar(&f.Category, "vc", "22", "set video category id")
	flag.StringVar(&f.PrivacyStatus, "video_privacystatus", "public", "set video privacy status")
	flag.StringVar(&f.PrivacyStatus, "vp", "public", "set video privacy status")
	flag.BoolVar(&f.Embeddable, "video_embeddable", true, "enable video to be embeddable")
	flag.BoolVar(&f.Embeddable, "ve", true, "enable video to be embeddable")
	flag.StringVar(&f.License, "video_license", "standard", "set video license")
	flag.StringVar(&f.License, "vl", "standard", "set video license")
	flag.BoolVar(&f.PublicStatsViewable, "video_publicstatsviewable", true, "enable public video stats to be viewable")
	flag.BoolVar(&f.PublicStatsViewable, "vs", true, "enable public video stats to be viewable")
	flag.StringVar(&f.PublishAt, "video_publishat", "", "set video publish time")
	flag.StringVar(&f.PublishAt, "vpa", "", "set video publish time")
	flag.StringVar(&f.RecordingDate, "video_recordingdate", "", "set video recording date")
	flag.StringVar(&f.RecordingDate, "vrd", "", "set video recording date")
	flag.StringVar(&f.PlaylistIds, "video_playlistids", "", "set video playlist ids")
	flag.StringVar(&f.PlaylistIds, "vpi", "", "set video playlist ids")
	flag.StringVar(&f.PlaylistTitles, "video_playlisttitles", "", "set video playlist titles")
	flag.StringVar(&f.PlaylistTitles, "vpt", "", "set video playlist titles")
	flag.Float64Var(&f.LocationLatitude, "video_location_latitude", 0, "set video latitude coordinate")
	flag.Float64Var(&f.LocationLatitude, "vla", 0, "set video latitude coordinate")
	flag.Float64Var(&f.LocationLongitude, "video_location_longitude", 0, "set video longitude coordinate")
	flag.Float64Var(&f.LocationLongitude, "vlo", 0, "set video longitude coordinate")
	flag.StringVar(&f.LocationDescription, "video_locationdescription", "", "set video location description")
	flag.StringVar(&f.LocationDescription, "vld", "", "set video location description")
	flag.IntVar(&f.UploadChunk, "upload_chunk", 8388608, "set upload chunk size in bytes")
	flag.IntVar(&f.UploadChunk, "uc", 8388608, "set upload chunk size in bytes")
	flag.Float64Var(&f.UploadRate, "upload_rate", 0, "set upload rate limit in kbps")
	flag.Float64Var(&f.UploadRate, "ur", 0, "set upload rate limit in kbps")
	flag.StringVar(&f.UploadTime, "upload_time", "", "set upload time limit ex- \"10:00-14:00\"")
	flag.StringVar(&f.UploadTime, "ut", "", "set upload time limit ex- \"10:00-14:00\"")
	flag.IntVar(&f.AuthPort, "auth_port", 8080, "set OAuth request port")
	flag.IntVar(&f.AuthPort, "ap", 8080, "set OAuth request port")
	flag.BoolVar(&f.AuthHeadless, "auth_headless", false, "enable browserless OAuth process")
	flag.BoolVar(&f.AuthHeadless, "ah", false, "enable browserless OAuth process")
	flag.BoolVar(&f.Version, "version", false, "show version")
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
	if f.Version {
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

// Search video by title (exact text)
func searchTitle(service *youtube.Service, text *string) {
	call, err := service.Search.List("snippet").Type("video").Q(*text).Do()
	if err != nil {
		if call != nil {
			log.Fatalf("Error searching video title  '%v': %v, %v", text, err, call.HTTPStatusCode)
		} else {
			log.Fatalf("Error searching video title '%v': %v", text, err)
		}
	}
	var re = regexp.MustCompile("\\W")
	var texta = strings.ToLower(re.ReplaceAllString(*text, ""))
	for _, item := range call.Items {
		var textb = strings.ToLower(re.ReplaceAllString(item.Snippet.Title, ""))
		if texta == textb {
			fmt.Printf("%v\n", item.Id.VideoId)
		}
	}
}

// Main.
func main() {
	// var videoFile io.ReadCloser
	// var fileSize int64
	var err error

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
	if f.Log {
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
		searchTitle(service, &f.Title)
		os.Exit(0)
	}
	if f.Title == "" {
		f.Title = "Video title"
	}

	if upload.Status.PrivacyStatus == "" {
		upload.Status.PrivacyStatus = f.PrivacyStatus
	}
	if upload.Snippet.Tags == nil && strings.Trim(f.Tags, "") != "" {
		upload.Snippet.Tags = strings.Split(f.Tags, ",")
	}
	if upload.Snippet.Title == "" {
		upload.Snippet.Title = f.Title
	}
	if upload.Snippet.Description == "" {
		upload.Snippet.Description = f.Description
	}
	if upload.Snippet.CategoryId == "" && f.Category != "" {
		upload.Snippet.CategoryId = f.Category
	}
	if upload.Snippet.DefaultLanguage == "" && f.Language != "" {
		upload.Snippet.DefaultLanguage = f.Language
	}
	if upload.Snippet.DefaultAudioLanguage == "" && f.Language != "" {
		upload.Snippet.DefaultAudioLanguage = f.Language
	}

	fmt.Printf("Uploading file '%s'...\n", f.Video)

	var option googleapi.MediaOption
	var video *youtube.Video

	option = googleapi.ChunkSize(f.UploadChunk)

	call := service.Videos.Insert("snippet,status,recordingDetails", upload)
	video, err = call.Media(videoFile, option).Do()

	if quitChan != nil {
		quit := make(chan struct{})
		quitChan <- quit
		<-quit
	}

	if err != nil {
		if video != nil {
			log.Fatalf("Error making YouTube API call: %v, %v", err, video.HTTPStatusCode)
		} else {
			log.Fatalf("Error making YouTube API call: %v", err)
		}
	}
	fmt.Printf("Upload successful! Video ID: %v\n", video.Id)

	if thumbnailFile != nil {
		log.Printf("Uploading thumbnail '%s'...\n", f.Thumbnail)
		_, err = service.Thumbnails.Set(video.Id).Media(thumbnailFile).Do()
		if err != nil {
			log.Fatalf("Error making YouTube API call: %v", err)
		}
		fmt.Printf("Thumbnail uploaded!\n")
	}

	// Insert caption
	if captionFile != nil {
		captionObj := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{},
		}
		captionObj.Snippet.VideoId = video.Id
		captionObj.Snippet.Language = f.Language
		captionObj.Snippet.Name = f.Language
		captionInsert := service.Captions.Insert("snippet", captionObj).Sync(true)
		captionRes, err := captionInsert.Media(captionFile).Do()
		if err != nil {
			if captionRes != nil {
				log.Fatalf("Error inserting caption: %v, %v", err, captionRes.HTTPStatusCode)
			} else {
				log.Fatalf("Error inserting caption: %v", err)
			}
		}
		fmt.Printf("Caption uploaded!\n")
	}

	plx := &Playlistx{}
	if upload.Status.PrivacyStatus != "" {
		plx.PrivacyStatus = upload.Status.PrivacyStatus
	}
	// PlaylistID is deprecated in favour of PlaylistIDs
	if videoMeta.PlaylistID != "" {
		plx.Id = videoMeta.PlaylistID
		err = plx.AddVideoToPlaylist(service, video.Id)
		if err != nil {
			log.Fatalf("Error adding video to playlist: %s", err)
		}
	}

	if len(videoMeta.PlaylistIDs) > 0 {
		plx.Title = ""
		for _, pid := range videoMeta.PlaylistIDs {
			plx.Id = pid
			err = plx.AddVideoToPlaylist(service, video.Id)
			if err != nil {
				log.Fatalf("Error adding video to playlist: %s", err)
			}
		}
	}

	if len(videoMeta.PlaylistTitles) > 0 {
		plx.Id = ""
		for _, title := range videoMeta.PlaylistTitles {
			plx.Title = title
			err = plx.AddVideoToPlaylist(service, video.Id)
			if err != nil {
				log.Fatalf("Error adding video to playlist: %s", err)
			}
		}
	}
}
