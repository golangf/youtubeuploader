package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

// Global variables
var appVersion = ""

func onTitle(srv *youtube.Service, txt string) {
	for id := range searchVideoTitle(srv, txt) {
		fmt.Printf("%v\n", id)
	}
}

// Main.
func main() {
	getFlags()
	// on help
	if f.Help {
		flag.PrintDefaults()
		os.Exit(1)
	}
	// on version
	if f.Version {
		fmt.Printf("youtubeuploader v%s\n", appVersion)
		os.Exit(0)
	}
	if f.Video == "" && f.Title == "" {
		fmt.Printf("No video file to upload!\n")
		os.Exit(1)
	}

	var id = f.Id
	uploadTime := getUploadTime()
	videoFile, fileSize := openFile(f.Video)
	thumbnailFile, _ := openFile(f.Thumbnail)
	captionFile, _ := openFile(f.Caption)
	if videoFile != nil {
		defer videoFile.Close()
	}
	if thumbnailFile != nil {
		defer thumbnailFile.Close()
	}
	if captionFile != nil {
		defer captionFile.Close()
	}

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
	// show video id
	if f.Video == "" && f.Id == "" && f.Title != "" {
		for _, id := range searchVideoTitle(service, f.Title) {
			fmt.Printf("%v\n", id)
		}
		os.Exit(0)
	}
	if f.PlaylistIds != "" && len(videoMeta.PlaylistIDs) == 0 {
		videoMeta.PlaylistIDs = strings.Split(f.PlaylistIds, ";")
	}
	if f.PlaylistTitles != "" && len(videoMeta.PlaylistTitles) == 0 {
		videoMeta.PlaylistTitles = strings.Split(f.PlaylistTitles, ";")
	}
	// update upload
	if id != "" || videoFile != nil {
		getUploadFlags(upload, &videoMeta)
		logUploadFlags(upload)
	}
	// DELETE ME!!!!!
	os.Exit(0)
	// upload video
	if videoFile != nil {
		logf("Uploading file '%s'...\n", f.Video)
		video := uploadVideo(service, videoFile, upload, parseInt(f.UploadChunk, 0), quitChan)
		logf("Upload successful! Video ID: %v\n", video.Id)
		id = video.Id
	} else if id != "" {
		logf("Updating video %v...\n", id)
		updateVideo(service, id, upload)
		logf("Update successful!\n")
	}
	// upload thumbnail
	if id != "" && thumbnailFile != nil {
		logf("Uploading thumbnail %v '%s'...\n", id, f.Thumbnail)
		uploadThumbnail(service, id, thumbnailFile)
		logf("Thumbnail uploaded!\n")
	}
	// upload caption
	if id != "" && captionFile != nil {
		logf("Uploading caption %v:%v '%s'...\n", id, f.Language, f.Caption)
		uploadCaption(service, id, captionFile)
		logf("Caption uploaded!\n")
	}
	// add to playlist id
	if id != "" && videoMeta.PlaylistID != "" {
		logf("Adding to playlist id %v->[%v]...\n", id, 1)
		addToPlaylistID(service, videoMeta.PlaylistID, upload.Status.PrivacyStatus, id)
	}
	// add to playlist ids
	if id != "" && videoMeta.PlaylistIDs != nil {
		logf("Adding to playlist ids %v->[%v]...\n", id, len(videoMeta.PlaylistIDs))
		addToPlaylistIDs(service, videoMeta.PlaylistIDs, upload.Status.PrivacyStatus, id)
	}
	// add to playlist titles
	if id != "" && videoMeta.PlaylistTitles != nil {
		logf("Adding to playlist ids %v->[%v]...\n", id, len(videoMeta.PlaylistTitles))
		addToPlaylistTitles(service, videoMeta.PlaylistTitles, upload.Status.PrivacyStatus, id)
	}
}
