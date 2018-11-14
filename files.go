package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/youtube/v3"
)

const ytDateLayout = "2006-01-02T15:04:05.000Z" // ISO 8601 (YYYY-MM-DDThh:mm:ss.sssZ)
const inputDateLayout = "2006-01-02"
const inputDatetimeLayout = "2006-01-02T15:04:05-07:00"

// Date defines a date
type Date struct {
	time.Time
}

// LoadVideoMeta loads metaJSON
func LoadVideoMeta(filename string, video *youtube.Video) (videoMeta VideoMeta) {
	// attempt to load from meta JSON, otherwise use values specified from command line flags
	if filename != "" {
		file, e := ioutil.ReadFile(filename)
		if e != nil {
			fmt.Printf("Error reading file '%s': %s\n", filename, e)
			fmt.Println("Will use command line flags instead")
			goto errJump
		}

		e = json.Unmarshal(file, &videoMeta)
		if e != nil {
			fmt.Printf("Error parsing file '%s': %s\n", filename, e)
			fmt.Println("Will use command line flags instead")
			goto errJump
		}

		video.Status = &youtube.VideoStatus{}
		video.Snippet.Tags = videoMeta.Tags
		video.Snippet.Title = videoMeta.Title
		video.Snippet.Description = videoMeta.Description
		video.Snippet.CategoryId = videoMeta.CategoryId
		if videoMeta.Location != nil {
			video.RecordingDetails.Location = videoMeta.Location
		}
		if videoMeta.LocationDescription != "" {
			video.RecordingDetails.LocationDescription = videoMeta.LocationDescription
		}
		if !videoMeta.RecordingDate.IsZero() {
			video.RecordingDetails.RecordingDate = videoMeta.RecordingDate.UTC().Format(ytDateLayout)
		}

		// status
		if videoMeta.PrivacyStatus != "" {
			video.Status.PrivacyStatus = videoMeta.PrivacyStatus
		}
		if videoMeta.Embeddable {
			video.Status.Embeddable = videoMeta.Embeddable
		}
		if videoMeta.License != "" {
			video.Status.License = videoMeta.License
		}
		if videoMeta.PublicStatsViewable {
			video.Status.PublicStatsViewable = videoMeta.PublicStatsViewable
		}
		if !videoMeta.PublishAt.IsZero() {
			if video.Status.PrivacyStatus != "private" {
				fmt.Printf("publishAt can only be used when privacyStatus is 'private'. Ignoring publishAt...\n")
			} else {
				if videoMeta.PublishAt.Before(time.Now()) {
					fmt.Printf("publishAt (%s) was in the past!? Publishing now instead...\n", videoMeta.PublishAt)
					video.Status.PublishAt = time.Now().UTC().Format(ytDateLayout)
				} else {
					video.Status.PublishAt = videoMeta.PublishAt.UTC().Format(ytDateLayout)
				}
			}
		}

		if videoMeta.Language != "" {
			video.Snippet.DefaultLanguage = videoMeta.Language
			video.Snippet.DefaultAudioLanguage = videoMeta.Language
		}
	}
errJump:

	if video.Status.PrivacyStatus == "" {
		video.Status.PrivacyStatus = f.PrivacyStatus
	}
	if video.Snippet.Tags == nil && strings.Trim(f.Tags, "") != "" {
		video.Snippet.Tags = strings.Split(f.Tags, ",")
	}
	if video.Snippet.Title == "" {
		video.Snippet.Title = f.Title
	}
	if video.Snippet.Description == "" {
		video.Snippet.Description = f.Description
	}
	if video.Snippet.CategoryId == "" && f.Category != "" {
		video.Snippet.CategoryId = f.Category
	}

	return
}

// Open opens a file
func Open(filename string) (io.ReadCloser, int64, error) {
	var reader io.ReadCloser
	var filesize int64
	var err error
	if strings.HasPrefix(filename, "http") {
		resp, err := http.Head(filename)
		if err != nil {
			return reader, filesize, fmt.Errorf("error opening %s: %s", filename, err)
		}
		lenStr := resp.Header.Get("content-length")
		if lenStr != "" {
			filesize, err = strconv.ParseInt(lenStr, 10, 64)
			if err != nil {
				return reader, filesize, err
			}
		}

		resp, err = http.Get(filename)
		if err != nil {
			return reader, filesize, fmt.Errorf("error opening %s: %s", filename, err)
		}
		if resp.ContentLength != 0 {
			filesize = resp.ContentLength
		}
		reader = resp.Body
		return reader, filesize, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return reader, filesize, fmt.Errorf("error opening %s: %s", filename, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return reader, filesize, fmt.Errorf("error stat'ing %s: %s", filename, err)
	}

	return file, fileInfo.Size(), nil
}

// UnmarshalJSON reads JSON
func (d *Date) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	s = s[1 : len(s)-1]
	// support ISO 8601 date only, and date + time
	if strings.ContainsAny(s, ":") {
		d.Time, err = time.Parse(inputDatetimeLayout, s)
	} else {
		d.Time, err = time.Parse(inputDateLayout, s)
	}
	return
}

func openFile(nam string) (io.ReadCloser, int64) {
	var fil io.ReadCloser
	var siz int64
	var err error
	if nam != "" {
		fil, siz, err = Open(nam)
		if err != nil {
			log.Fatal(err)
		}
	}
	return fil, siz
}
