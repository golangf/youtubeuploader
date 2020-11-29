package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/porjo/go-flowrate/flowrate"
	"google.golang.org/api/youtube/v3"
)

//
// Types
//
type limitTransport struct {
	rt       http.RoundTripper
	lr       limitRange
	reader   *flowrate.Reader
	filesize int64
}

type Playlistx struct {
	Id            string
	Title         string
	PrivacyStatus string
}

type VideoMeta struct {
	// snippet
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	CategoryId  string   `json:"categoryId,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// status
	PrivacyStatus       string `json:"privacyStatus,omitempty"`
	Embeddable          bool   `json:"embeddable,omitempty"`
	License             string `json:"license,omitempty"`
	PublicStatsViewable bool   `json:"publicStatsViewable,omitempty"`
	PublishAt           Date   `json:"publishAt,omitempty"`

	// recording details
	Location            *youtube.GeoPoint `json:"location,omitempty"`
	LocationDescription string            `json:"locationDescription,omitempty"`
	RecordingDate       Date              `json:"recordingDate,omitempty"`

	// PlaylistID is deprecated in favour of PlaylistIDs
	PlaylistID     string   `json:"playlistId,omitempty"`
	PlaylistIDs    []string `json:"playlistIds,omitempty"`
	PlaylistTitles []string `json:"playlistTitles,omitempty"`

	// BCP-47 language code e.g. 'en','es'
	Language string `json:"language,omitempty"`

	// JSON map
	JSON map[string]interface{}
}

//
// Global constants
//
const inputTimeLayout = "15:04"

//
// Functions
//
func (t *limitTransport) RoundTrip(r *http.Request) (res *http.Response, err error) {
	// Content-Type starts with 'multipart/related' where chunksize >= filesize (including chunksize 0)
	// and 'video' for other chunksizes
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/related") ||
		strings.HasPrefix(r.Header.Get("Content-Type"), "video") {
		var monitor *flowrate.Monitor

		if t.reader != nil {
			monitor = t.reader.Monitor
		}

		// limit is set in limitChecker.Read
		t.reader = flowrate.NewReader(r.Body, 0)

		if monitor != nil {
			// carry over stats to new limiter
			t.reader.Monitor = monitor
		} else {
			t.reader.Monitor.SetTransferSize(t.filesize)
		}
		r.Body = &limitChecker{t.lr, t.reader}
	}

	return t.rt.RoundTrip(r)
}

func (plx *Playlistx) AddVideoToPlaylist(service *youtube.Service, videoID string) (err error) {
	listCall := service.Playlists.List([]string{"snippet", "contentDetails"})
	listCall = listCall.Mine(true)
	response, err := listCall.Do()
	if err != nil {
		return fmt.Errorf("Error retrieving playlists: %s", err)
	}

	var playlist *youtube.Playlist
	for _, pl := range response.Items {
		if pl.Id == plx.Id || pl.Snippet.Title == plx.Title {
			playlist = pl
			break
		}
	}

	// create playlist if it doesn't exist
	if playlist == nil {
		if plx.Id != "" {
			return fmt.Errorf("Playlist ID '%s' doesn't exist", plx.Id)
		}
		playlist = &youtube.Playlist{}
		playlist.Snippet = &youtube.PlaylistSnippet{Title: plx.Title}
		playlist.Status = &youtube.PlaylistStatus{PrivacyStatus: plx.PrivacyStatus}
		insertCall := service.Playlists.Insert([]string{"snippet", "status"}, playlist)
		// API doesn't return playlist ID here!?
		playlist, err = insertCall.Do()
		if err != nil {
			return fmt.Errorf("Error creating playlist with title '%s': %s", plx.Title, err)
		}
	}

	playlistItem := &youtube.PlaylistItem{}
	playlistItem.Snippet = &youtube.PlaylistItemSnippet{PlaylistId: playlist.Id, Title: playlist.Snippet.Title}
	playlistItem.Snippet.ResourceId = &youtube.ResourceId{
		VideoId: videoID,
		Kind:    "youtube#video",
	}

	insertCall := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err = insertCall.Do()
	if err != nil {
		return err
	}

	logf("Video added to playlist '%s' (%s)\n", playlist.Snippet.Title, playlist.Id)

	return nil
}
