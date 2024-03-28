package parasect_test

import (
	"testing"

	"github.com/nanoteck137/parasect"
)

func TestParseTrackName(t *testing.T) {
	tests := []struct {
		name           string
		trackName      string
		expectedName   string
		expectedNumber int
	}{
		{
			name:           "normal track name",
			trackName:      "01. track.flac",
			expectedName:   "track",
			expectedNumber: 1,
		},
		{
			name:           "spaces in name",
			trackName:      "02 some track name.flac",
			expectedName:   "some track name",
			expectedNumber: 2,
		},
		{
			name:           "no name",
			trackName:      "03.flac",
			expectedName:   "03.flac",
			expectedNumber: 3,
		},
		{
			name:           "no name and number",
			trackName:      "10.flac",
			expectedName:   "10.flac",
			expectedNumber: 10,
		},
		{
			name:           "dash seperator",
			trackName:      "23 - hello world.flac",
			expectedName:   "hello world",
			expectedNumber: 23,
		},
		{
			name:           "no spaces between seperator (dash)",
			trackName:      "100-hello world.flac",
			expectedName:   "hello world",
			expectedNumber: 100,
		},
		{
			name:           "no spaces between seperator (dot)",
			trackName:      "124.hello world.flac",
			expectedName:   "hello world",
			expectedNumber: 124,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tn, err := parasect.ParseTrackName(test.trackName)
			if err != nil {
				t.Fatalf("%s: got error while parsing track name: %#v", test.name, err)
			}

			if tn.Name != test.expectedName {
				t.Fatalf("%s: expected name: %#v, but got: %#v", test.name, test.expectedName, tn.Name)
			}

			if tn.Number != test.expectedNumber {
				t.Fatalf("%s: expected number: %#v, but got: %#v", test.name, test.expectedNumber, tn.Number)
			}
		})
	}
}
