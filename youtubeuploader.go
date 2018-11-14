package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

// Global variables
var appVersion = ""

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

// Main.
func main() {
	getFlags()
	onVersion(&f)
	onHelp(&f)
	uploadTime := getUploadTime()
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
	updateMeta(upload, &f)
	video := uploadVideo(service, f.Video, videoFile, upload, parseInt(f.UploadChunk, 0), quitChan)
	uploadThumbnail(service, video.Id, f.Thumbnail, thumbnailFile)
	uploadCaption(service, video.Id, f.Caption, captionFile)
	addToPlaylistID(service, videoMeta.PlaylistID, upload.Status.PrivacyStatus, video.Id)
	addToPlaylistIDs(service, videoMeta.PlaylistIDs, upload.Status.PrivacyStatus, video.Id)
	addToPlaylistTitles(service, videoMeta.PlaylistTitles, upload.Status.PrivacyStatus, video.Id)
}
