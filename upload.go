package main

import (
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

//
// Global variables
//
var reTemplate = regexp.MustCompile("\\$\\{.*?\\}")

//
// Functions
//
// Limit title to 100 characters.
func limitTitle(txt string) string {
	var re = regexp.MustCompile("<|>")
	txt = re.ReplaceAllString(txt, "")
	if len(txt) <= 100 {
		return txt
	}
	return txt[0:96] + " ..."
}

// Limit description to 5000 characters.
func limitDescription(txt string) string {
	var re = regexp.MustCompile("<|>")
	txt = re.ReplaceAllString(txt, "")
	if len(txt) <= 5000 {
		return txt
	}
	return txt[0:4996] + " ..."
}

// Limit tags to 30, 450 characters.
func limitTags(tags []string) []string {
	var l = 0
	var z []string
	for _, tag := range tags {
		var t = strings.ToLower(tag)
		if len(t) > 30 || stringsIncludes(z, t) {
			continue
		}
		if l+len(t)+2 > 450 {
			break
		}
		z = append(z, t)
		l += len(t) + 2
	}
	return z
}

func getUploadFlagsDefault(y *youtube.Video) {
	y.Snippet.Title = parseString(y.Snippet.Title, f.Video)
	y.Snippet.Description = parseString(y.Snippet.Description, f.Video)
	y.Snippet.DefaultLanguage = parseString(y.Snippet.DefaultLanguage, "en")
	y.Snippet.DefaultAudioLanguage = parseString(y.Snippet.DefaultAudioLanguage, "en")
	y.Snippet.CategoryId = parseString(y.Snippet.CategoryId, "22")
	y.Status.PrivacyStatus = parseString(y.Status.PrivacyStatus, "private")
}

func getUploadFlagsDynamic(y *youtube.Video, m *VideoMeta) {
	if y.Snippet.Tags == nil {
		y.Snippet.Tags = []string{}
	}
	if f.Title != "" {
		y.Snippet.Title = mapString(f.Title, m.JSON)
	}
	if f.Description != "" {
		y.Snippet.Description = mapString(f.Description, m.JSON)
	}
	if f.Tags != "" {
		y.Snippet.Tags = strings.Split(mapString(f.Tags, m.JSON), ",")
	}
	if f.Category != "" {
		y.Snippet.CategoryId = strconv.Itoa(parseCategory(f.Category))
	}
	if f.License != "" {
		y.Status.License = parseLicense(f.License)
	}
	if f.PublicStatsViewable {
		y.Status.PublicStatsViewable = f.PublicStatsViewable
	}
	if f.LocationLatitude != "" || f.LocationLongitude != "" {
		if y.RecordingDetails.Location == nil {
			y.RecordingDetails.Location = &youtube.GeoPoint{}
		}
		y.RecordingDetails.Location.Latitude = parseFloat(f.LocationLatitude, y.RecordingDetails.Location.Latitude)
		y.RecordingDetails.Location.Longitude = parseFloat(f.LocationLongitude, y.RecordingDetails.Location.Longitude)
	}
	y.Status.PrivacyStatus = parseString(f.PrivacyStatus, y.Status.PrivacyStatus)
	y.Status.PublishAt = parseString(f.PublishAt, y.Status.PublishAt)
	y.RecordingDetails.RecordingDate = parseString(f.RecordingDate, y.RecordingDetails.RecordingDate)
	y.RecordingDetails.LocationDescription = parseString(f.LocationDescription, y.RecordingDetails.LocationDescription)
}

func getUploadFlags(y *youtube.Video, m *VideoMeta) {
	getUploadFlagsDynamic(y, m)
	getUploadFlagsDefault(y)
	y.Snippet.Title = limitTitle(y.Snippet.Title)
	y.Snippet.Description = limitDescription(y.Snippet.Description)
	y.Snippet.Tags = limitTags(y.Snippet.Tags)
}

func searchVideoTitle(srv *youtube.Service, txt string) []string {
	res, err := srv.Search.List("snippet").Type("video").Q(txt).Do()
	if err != nil {
		if res != nil {
			log.Fatalf("Error searching video title  '%v': %v, %v", txt, err, res.HTTPStatusCode)
		} else {
			log.Fatalf("Error searching video title '%v': %v", txt, err)
		}
	}
	var ans = []string{}
	var re = regexp.MustCompile("\\W")
	var ta = strings.ToLower(re.ReplaceAllString(txt, ""))
	for _, item := range res.Items {
		var tb = strings.ToLower(re.ReplaceAllString(item.Snippet.Title, ""))
		if ta == tb {
			ans = append(ans, item.Id.VideoId)
		}
	}
	return ans
}

func updateVideo(srv *youtube.Service, id string, obj *youtube.Video) {
	obj.Id = id
	res, err := srv.Videos.Update("snippet,status,recordingDetails", obj).Do()
	if err != nil {
		if res != nil {
			log.Fatalf("Error updating video: %v, %v", err, res.HTTPStatusCode)
		} else {
			log.Fatalf("Error updating video: %v", err)
		}
	}
}

func uploadVideo(srv *youtube.Service, fil io.ReadCloser, obj *youtube.Video, cnk int, cquit chanChan) *youtube.Video {
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
	return res
}

func uploadThumbnail(srv *youtube.Service, id string, fil io.ReadCloser) {
	res, err := srv.Thumbnails.Set(id).Media(fil).Do()
	if err != nil {
		if res != nil {
			log.Fatalf("Error uploading thumbnail: %v, %v", err, res.HTTPStatusCode)
		} else {
			log.Fatalf("Error uploading thumbnail: %v", err)
		}
	}
}

func uploadCaption(srv *youtube.Service, id string, lng string, fil io.ReadCloser) {
	c := &youtube.Caption{
		Snippet: &youtube.CaptionSnippet{},
	}
	c.Snippet.VideoId = id
	c.Snippet.Language = lng
	c.Snippet.Name = lng
	req := srv.Captions.Insert("snippet", c).Sync(true)
	res, err := req.Media(fil).Do()
	if err != nil {
		if res != nil {
			log.Fatalf("Error uploading caption: %v, %v", err, res.HTTPStatusCode)
		} else {
			log.Fatalf("Error uploading caption: %v", err)
		}
	}
}

func addToPlaylistID(srv *youtube.Service, pid string, sta string, id string) {
	p := Playlistx{}
	p.PrivacyStatus = sta
	// PlaylistID is deprecated in favour of PlaylistIDs
	p.Id = pid
	err := p.AddVideoToPlaylist(srv, id)
	if err != nil {
		log.Fatalf("Error adding video to playlist: %s", err)
	}
}

func addToPlaylistIDs(srv *youtube.Service, pids []string, sta string, id string) {
	p := Playlistx{}
	p.PrivacyStatus = sta
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
