package main

import "google.golang.org/api/youtube/v3"
import "google.golang.org/api/googleapi"
import "golang.org/x/oauth2"
import "net/http"
import "strings"
import "context"
import "regexp"
import "flag"
import "fmt"
import "log"
import "io"
import "os"

type chanChan chan chan struct{}

func main2() {
	var video = flag.String("video", "", "set input video file")
	flag.StringVar(video, "v", "", "set input video file")
	var thumbnail = flag.String("thumbnail", "", "set input thumbnail file")
	flag.StringVar(thumbnail, "t", "", "set input thumbnail file")
	var caption = flag.String("caption", "", "set input caption file")
	flag.StringVar(caption, "c", "", "set input caption file")
	var meta = flag.String("meta", "", "set input meta file")
	flag.StringVar(meta, "m", "", "set input meta file")
	var log = flag.Bool("log", false, "enable log")
	flag.BoolVar(log, "l", false, "enable log")
	var clientID = flag.String("client_id", "client_id.json", "set client id credentials path")
	flag.StringVar(clientID, "ci", "client_id.json", "set client id credentials path")
	var clientToken = flag.String("client_token", "client_token.json", "set client token credentials path")
	flag.StringVar(clientToken, "ct", "client_token.json", "set client token credentials path")
	var videoTitle = flag.String("video_title", "", "set video title")
	flag.StringVar(videoTitle, "vt", "", "set video title")
	var videoDescription = flag.String("video_description", "", "set video description")
	flag.StringVar(videoDescription, "vd", "", "set video description")
	var videoTags = flag.String("video_tags", "", "set video tags/keywords")
	flag.StringVar(videoTags, "vk", "", "set video tags/keywords")
	var videoLanguage = flag.String("video_language", "en", "set video language")
	flag.StringVar(videoLanguage, "vl", "en", "set video language")
	var videoCategory = flag.String("video_category", "22", "set video category id")
	flag.StringVar(videoCategory, "vc", "22", "set video category id")
	var videoPrivacyStatus = flag.String("video_privacystatus", "public", "set video privacy status")
	flag.StringVar(videoPrivacyStatus, "vp", "public", "set video privacy status")
	var videoEmbeddable = flag.Bool("video_embeddable", true, "enable video to be embeddable")
	flag.BoolVar(videoEmbeddable, "ve", true, "enable video to be embeddable")
	var videoLicense = flag.String("video_license", "standard", "set video license")
	flag.StringVar(videoLicense, "vl", "standard", "set video license")
	var videoPublicStatsViewable = flag.Bool("video_publicstatsviewable", true, "enable public video stats to be viewable")
	flag.BoolVar(videoPublicStatsViewable, "vs", true, "enable public video stats to be viewable")
	var videoPublishAt = flag.String("video_publishat", "", "set video publish time")
	flag.StringVar(videoPublishAt, "vpa", "", "set video publish time")
	var videoRecordingDate = flag.String("video_recordingdate", "", "set video recording date")
	flag.StringVar(videoRecordingDate, "vrd", "", "set video recording date")
	var videoPlaylistIds = flag.String("video_playlistids", "", "set video playlist ids")
	flag.StringVar(videoPlaylistIds, "vpi", "", "set video playlist ids")
	var videoPlaylistTitles = flag.String("video_playlisttitles", "", "set video playlist titles")
	flag.StringVar(videoPlaylistTitles, "vpt", "", "set video playlist titles")
	var videoLocationLatitude = flag.Float64("video_location_latitude", 0, "set video latitude coordinate")
	flag.Float64Var(videoLocationLatitude, "vla", 0, "set video latitude coordinate")
	var videoLocationLongitude = flag.Float64("video_location_longitude", 0, "set video longitude coordinate")
	flag.Float64Var(videoLocationLongitude, "vlo", 0, "set video longitude coordinate")
	var videoLocationDescription = flag.String("video_locationdescription", "", "set video location description")
	flag.StringVar(videoLocationDescription, "vld", "", "set video location description")
	var uploadChunk = flag.Int("upload_chunk", 8388608, "set upload chunk size in bytes")
	flag.IntVar(uploadChunk, "uc", 8388608, "set upload chunk size in bytes")
	var uploadRate = flag.Float64("upload_rate", 0, "set upload rate limit in kbps")
	flag.Float64Var(uploadRate, "ur", 0, "set upload rate limit in kbps")
	var uploadTime = flag.String("upload_time", "", "set upload time limit ex- \"10:00-14:00\"")
	flag.StringVar(uploadTime, "ut", "", "set upload time limit ex- \"10:00-14:00\"")
	var authPort = flag.Int("auth_port", 8080, "set OAuth request port")
	flag.IntVar(authPort, "ap", 8080, "set OAuth request port")
	var authHeadless = flag.Bool("auth_headless", false, "enable browserless OAuth process")
	flag.BoolVar(authHeadless, "ah", false, "enable browserless OAuth process")
}

var (
	filename       = flag.String("filename", "", "Filename to upload. Can be a URL")
	thumbnail      = flag.String("thumbnail", "", "Thumbnail to upload. Can be a URL")
	caption        = flag.String("caption", "", "Caption to upload. Can be URL")
	title          = flag.String("title", "", "Video title")
	description    = flag.String("description", "uploaded by youtubeuploader", "Video description")
	language       = flag.String("language", "en", "Video language")
	categoryId     = flag.String("categoryId", "", "Video category Id")
	tags           = flag.String("tags", "", "Comma separated list of video tags")
	privacy        = flag.String("privacy", "private", "Video privacy status")
	quiet          = flag.Bool("quiet", false, "Suppress progress indicator")
	rate           = flag.Int("ratelimit", 0, "Rate limit upload in kbps. No limit by default")
	metaJSON       = flag.String("metaJSON", "", "JSON file containing title,description,tags etc (optional)")
	limitBetween   = flag.String("limitBetween", "", "Only rate limit between these times e.g. 10:00-14:00 (local time zone)")
	headlessAuth   = flag.Bool("headlessAuth", false, "set this if no browser available for the oauth authorisation step")
	oAuthPort      = flag.Int("oAuthPort", 8080, "TCP port to listen on when requesting an oAuth token")
	showAppVersion = flag.Bool("v", false, "show version")
	chunksize      = flag.Int("chunksize", googleapi.DefaultUploadChunkSize, "size (in bytes) of each upload chunk. A zero value will cause all data to be uploaded in a single request")

	// this is set by compile-time to match git tag
	appVersion string = "unknown"
)

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
	flag.Parse()

	if *showAppVersion {
		fmt.Printf("Youtubeuploader version: %s\n", appVersion)
		os.Exit(0)
	}

	if *filename == "" && *title == "" {
		fmt.Printf("You must provide a filename of a video file to upload\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var reader io.ReadCloser
	var filesize int64
	var err error

	var limitRange limitRange
	if *limitBetween != "" {
		limitRange, err = parseLimitBetween(*limitBetween)
		if err != nil {
			fmt.Printf("Invalid value for -limitBetween: %v", err)
			os.Exit(1)
		}
	}

	if *filename != "" {
		reader, filesize, err = Open(*filename)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
	}

	var thumbReader io.ReadCloser
	if *thumbnail != "" {
		thumbReader, _, err = Open(*thumbnail)
		if err != nil {
			log.Fatal(err)
		}
		defer thumbReader.Close()
	}

	var captionReader io.ReadCloser
	if *caption != "" {
		captionReader, _, err = Open(*caption)
		if err != nil {
			log.Fatal(err)
		}
		defer captionReader.Close()
	}

	ctx := context.Background()
	transport := &limitTransport{rt: http.DefaultTransport, lr: limitRange, filesize: filesize}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: transport,
	})

	var quitChan chanChan
	if !*quiet {
		quitChan = make(chanChan)
		go func() {
			Progress(quitChan, transport, filesize)
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

	videoMeta := LoadVideoMeta(*metaJSON, upload)

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating Youtube client: %s", err)
	}

	// Search video by title
	if *filename == "" && *title != "" {
		searchTitle(service, title)
		os.Exit(0)
	}
	if *title == "" {
		*title = "Video title"
	}

	if upload.Status.PrivacyStatus == "" {
		upload.Status.PrivacyStatus = *privacy
	}
	if upload.Snippet.Tags == nil && strings.Trim(*tags, "") != "" {
		upload.Snippet.Tags = strings.Split(*tags, ",")
	}
	if upload.Snippet.Title == "" {
		upload.Snippet.Title = *title
	}
	if upload.Snippet.Description == "" {
		upload.Snippet.Description = *description
	}
	if upload.Snippet.CategoryId == "" && *categoryId != "" {
		upload.Snippet.CategoryId = *categoryId
	}
	if upload.Snippet.DefaultLanguage == "" && *language != "" {
		upload.Snippet.DefaultLanguage = *language
	}
	if upload.Snippet.DefaultAudioLanguage == "" && *language != "" {
		upload.Snippet.DefaultAudioLanguage = *language
	}

	fmt.Printf("Uploading file '%s'...\n", *filename)

	var option googleapi.MediaOption
	var video *youtube.Video

	option = googleapi.ChunkSize(*chunksize)

	call := service.Videos.Insert("snippet,status,recordingDetails", upload)
	video, err = call.Media(reader, option).Do()

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

	if thumbReader != nil {
		log.Printf("Uploading thumbnail '%s'...\n", *thumbnail)
		_, err = service.Thumbnails.Set(video.Id).Media(thumbReader).Do()
		if err != nil {
			log.Fatalf("Error making YouTube API call: %v", err)
		}
		fmt.Printf("Thumbnail uploaded!\n")
	}

	// Insert caption
	if captionReader != nil {
		captionObj := &youtube.Caption{
			Snippet: &youtube.CaptionSnippet{},
		}
		captionObj.Snippet.VideoId = video.Id
		captionObj.Snippet.Language = *language
		captionObj.Snippet.Name = *language
		captionInsert := service.Captions.Insert("snippet", captionObj).Sync(true)
		captionRes, err := captionInsert.Media(captionReader).Do()
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
