package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

//
// Functions
//
func updateMeta(m *youtube.Video) {
	if m.Snippet.Title == "" {
		m.Snippet.Title = f.Title
	}
	if m.Snippet.Title == "" {
		m.Snippet.Title = f.Video
	}
	if m.Snippet.Description == "" && f.DescriptionPath != "" {
		dat, err := ioutil.ReadFile(f.DescriptionPath)
		if err != nil {
			log.Fatalf("Error reading description file '%v': %v", f.DescriptionPath, err)
		}
		m.Snippet.Description = string(dat)
	}
	if m.Snippet.Description == "" {
		m.Snippet.Description = f.Description
	}
	if m.Snippet.Tags == nil && strings.Trim(f.Tags, "") != "" {
		m.Snippet.Tags = strings.Split(f.Tags, ",")
	}
	if m.Snippet.DefaultLanguage == "" {
		m.Snippet.DefaultLanguage = f.Language
	}
	if m.Snippet.DefaultAudioLanguage == "" {
		m.Snippet.DefaultAudioLanguage = f.Language
	}
	if m.Snippet.CategoryId == "" {
		m.Snippet.CategoryId = string(parseCategory(f.Category))
	}
	if m.Status.PrivacyStatus == "" {
		m.Status.PrivacyStatus = f.PrivacyStatus
	}
	if m.Status.License == "" {
		m.Status.License = parseLicense(f.License)
	}
	if !m.Status.PublicStatsViewable {
		m.Status.PublicStatsViewable = f.PublicStatsViewable
	}
	if m.Status.PublishAt == "" {
		m.Status.PublishAt = f.PublishAt
	}
	if m.RecordingDetails.RecordingDate == "" {
		m.RecordingDetails.RecordingDate = f.RecordingDate
	}
	if math.IsNaN(m.RecordingDetails.Location.Latitude) {
		m.RecordingDetails.Location.Latitude = parseFloat(f.LocationLatitude, math.NaN())
	}
	if math.IsNaN(m.RecordingDetails.Location.Longitude) {
		m.RecordingDetails.Location.Longitude = parseFloat(f.LocationLongitude, math.NaN())
	}
	if m.RecordingDetails.LocationDescription == "" {
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
