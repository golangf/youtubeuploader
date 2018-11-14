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
func LoadVideoMeta(filename string, y *youtube.Video) (m VideoMeta) {
	// attempt to load from meta JSON, otherwise use values specified from command line flags
	if filename != "" {
		file, e := ioutil.ReadFile(filename)
		if e != nil {
			fmt.Printf("Error reading file '%s': %s\n", filename, e)
			fmt.Println("Will use command line flags instead")
			goto errJump
		}

		e = json.Unmarshal(file, &m)
		if e != nil {
			fmt.Printf("Error parsing file '%s': %s\n", filename, e)
			fmt.Println("Will use command line flags instead")
			goto errJump
		}

		y.Status = &youtube.VideoStatus{}
		y.Snippet.Tags = m.Tags
		y.Snippet.Title = m.Title
		y.Snippet.Description = m.Description
		y.Snippet.CategoryId = m.CategoryId
		if m.Location != nil {
			y.RecordingDetails.Location = m.Location
		}
		if m.LocationDescription != "" {
			y.RecordingDetails.LocationDescription = m.LocationDescription
		}
		if !m.RecordingDate.IsZero() {
			y.RecordingDetails.RecordingDate = m.RecordingDate.UTC().Format(ytDateLayout)
		}

		// status
		if m.PrivacyStatus != "" {
			y.Status.PrivacyStatus = m.PrivacyStatus
		}
		if m.Embeddable {
			y.Status.Embeddable = m.Embeddable
		}
		if m.License != "" {
			y.Status.License = m.License
		}
		if m.PublicStatsViewable {
			y.Status.PublicStatsViewable = m.PublicStatsViewable
		}
		if !m.PublishAt.IsZero() {
			if y.Status.PrivacyStatus != "private" {
				fmt.Printf("publishAt can only be used when privacyStatus is 'private'. Ignoring publishAt...\n")
			} else {
				if m.PublishAt.Before(time.Now()) {
					fmt.Printf("publishAt (%s) was in the past!? Publishing now instead...\n", m.PublishAt)
					y.Status.PublishAt = time.Now().UTC().Format(ytDateLayout)
				} else {
					y.Status.PublishAt = m.PublishAt.UTC().Format(ytDateLayout)
				}
			}
		}
		if m.Language != "" {
			y.Snippet.DefaultLanguage = m.Language
			y.Snippet.DefaultAudioLanguage = m.Language
		}
	}
errJump:
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
