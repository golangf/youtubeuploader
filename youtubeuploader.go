package main

import "google.golang.org/api/youtube/v3"
import "google.golang.org/api/googleapi"
import "golang.org/x/oauth2"
import "net/http"

import "strings"
import "context"
import "regexp"
import "flag"
import "math"
import "fmt"
import "log"
import "io"
import "os"

// Global variables
var appVersion = ""

func openFile(nam string) (io.ReadCloser, int64) {
	var fil io.ReadCloser
	var siz int64
	var err error
	if nam != "" {
		fil, siz, err = Open(nam)
		if err != nil {
			log.Fatal(err)
		}
		defer fil.Close()
	}
	return fil, siz
}

func getUploadTime(f *appFlags) limitRange {
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

func onVersion(f *appFlags) {
	if f.Version {
		fmt.Printf("youtubeuploader v%s\n", appVersion)
		os.Exit(0)
	}
}

func onHelp(f *appFlags) {
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

func updateFlags(f *appFlags) {
	if f.Title == "" {
		f.Title = f.Video
	}
	if f.Description == "" {
		f.Description = f.Video
	}
}

func updateMeta(m *youtube.Video, f *appFlags) {
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
	if f.PublicStatsViewable {
		m.Status.PublicStatsViewable = f.PublicStatsViewable
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
	getFlags()
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
