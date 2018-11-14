package main

import (
	"io"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

//
// Functions
//
func updateUpload(y *youtube.Video) {
	if y.Snippet.Title == "" {
		y.Snippet.Title = f.Title
	}
	if y.Snippet.Title == "" {
		y.Snippet.Title = f.Video
	}
	if y.Snippet.Description == "" && f.DescriptionPath != "" {
		dat, err := ioutil.ReadFile(f.DescriptionPath)
		if err != nil {
			log.Fatalf("Error reading description file '%v': %v", f.DescriptionPath, err)
		}
		y.Snippet.Description = string(dat)
	}
	if y.Snippet.Description == "" {
		y.Snippet.Description = f.Description
	}
	if y.Snippet.Tags == nil && strings.Trim(f.Tags, "") != "" {
		y.Snippet.Tags = strings.Split(f.Tags, ",")
	}
	if y.Snippet.DefaultLanguage == "" {
		y.Snippet.DefaultLanguage = f.Language
	}
	if y.Snippet.DefaultAudioLanguage == "" {
		y.Snippet.DefaultAudioLanguage = f.Language
	}
	if y.Snippet.CategoryId == "" {
		y.Snippet.CategoryId = strconv.Itoa(parseCategory(f.Category))
	}
	if y.Status.PrivacyStatus == "" {
		y.Status.PrivacyStatus = f.PrivacyStatus
	}
	if y.Status.License == "" {
		y.Status.License = parseLicense(f.License)
	}
	if !y.Status.PublicStatsViewable {
		y.Status.PublicStatsViewable = f.PublicStatsViewable
	}
	if y.Status.PublishAt == "" {
		y.Status.PublishAt = f.PublishAt
	}
	if y.RecordingDetails.RecordingDate == "" {
		y.RecordingDetails.RecordingDate = f.RecordingDate
	}
	if f.LocationLatitude != "" && f.LocationLongitude != "" {
		if y.RecordingDetails.Location == nil {
			y.RecordingDetails.Location = &youtube.GeoPoint{}
		}
		if math.IsNaN(y.RecordingDetails.Location.Latitude) {
			y.RecordingDetails.Location.Latitude = parseFloat(f.LocationLatitude, math.NaN())
		}
		if math.IsNaN(y.RecordingDetails.Location.Longitude) {
			y.RecordingDetails.Location.Longitude = parseFloat(f.LocationLongitude, math.NaN())
		}
	}
	if y.RecordingDetails.LocationDescription == "" {
		y.RecordingDetails.LocationDescription = f.LocationDescription
	}
}

func uploadVideo(srv *youtube.Service, nam string, fil io.ReadCloser, obj *youtube.Video, cnk int, cquit chanChan) *youtube.Video {
	logf("Uploading file '%s'...\n", nam)
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
	logf("Upload successful! Video ID: %v\n", res.Id)
	return res
}

func uploadThumbnail(srv *youtube.Service, id string, nam string, fil io.ReadCloser) {
	if fil != nil {
		logf("Uploading thumbnail '%s'...\n", nam)
		res, err := srv.Thumbnails.Set(id).Media(fil).Do()
		if err != nil {
			if res != nil {
				log.Fatalf("Error uploading thumbnail: %v, %v", err, res.HTTPStatusCode)
			} else {
				log.Fatalf("Error uploading thumbnail: %v", err)
			}
		}
		logf("Thumbnail uploaded!\n")
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
		logf("Uploading caption '%s'...\n", nam)
		req := srv.Captions.Insert("snippet", c).Sync(true)
		res, err := req.Media(fil).Do()
		if err != nil {
			if res != nil {
				log.Fatalf("Error uploading caption: %v, %v", err, res.HTTPStatusCode)
			} else {
				log.Fatalf("Error uploading caption: %v", err)
			}
		}
		logf("Caption uploaded!\n")
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
