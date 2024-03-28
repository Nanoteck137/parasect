package parasect

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func RunFFprobe(verbose bool, args ...string) ([]byte, error) {
	cmd := exec.Command("ffprobe", args...)
	if verbose {
		cmd.Stderr = os.Stderr
	}

	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func RunFFmpeg(verbose bool, args ...string) error {
	cmd := exec.Command("ffmpeg", args...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

type TrackInfo struct {
	Path   string
	Number int
	Name   string

	Duration int
	Tags     map[string]string
}

type probeFormat struct {
	BitRate    string            `json:"bit_rate"`
	Tags       map[string]string `json:"tags"`
	Duration   string            `json:"duration"`
	FormatName string            `json:"format_name"`
}

type probeStream struct {
	Index     int    `json:"index"`
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`

	Duration string `json:"duration"`

	Tags map[string]string `json:"tags"`
}

type probe struct {
	Streams []probeStream `json:"streams"`
	Format  probeFormat   `json:"format"`
}

// TODO(patrik): Test
func getNumberFromFormatString(s string) int {
	if strings.Contains(s, "/") {
		s = strings.Split(s, "/")[0]
	}

	num, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}

	return num
}

// TODO(patrik): Test
func convertMapKeysToLowercase(m map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		res[strings.ToLower(k)] = v
	}

	return res
}

type ProbeResult struct {
	Tags     map[string]string
	Duration int
}

func ProbeTrack(filepath string) (ProbeResult, error) {
	data, err := RunFFprobe(false, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filepath)
	if err != nil {
		return ProbeResult{}, err
	}

	var probe probe
	err = json.Unmarshal(data, &probe)
	if err != nil {
		return ProbeResult{}, err
	}

	hasGlobalTags := probe.Format.FormatName != "ogg"

	var tags map[string]string

	if hasGlobalTags {
		tags = convertMapKeysToLowercase(probe.Format.Tags)
	}

	duration := 0
	for _, s := range probe.Streams {
		if s.CodecType == "audio" {
			dur, err := strconv.ParseFloat(s.Duration, 32)
			if err != nil {
				return ProbeResult{}, err
			}

			duration = int(dur)
			if !hasGlobalTags {
				tags = convertMapKeysToLowercase(s.Tags)
			}
		}
	}

	return ProbeResult{
		Tags:     tags,
		Duration: duration,
	}, nil
}

var trackNameRegex = regexp.MustCompile(`(^\d+)[-\s.]*(.+)?\.`)

type TrackName struct {
	Name   string
	Number int
}

func ParseTrackName(n string) (TrackName, error) {
	res := trackNameRegex.FindStringSubmatch(n)
	num, err := strconv.Atoi(res[1])
	if err != nil {
		return TrackName{}, fmt.Errorf("failed ParseTrackName: %w", err)
	}

	name := res[2]
	if name == "" {
		name = n
	}

	return TrackName{
		Name:   name,
		Number: num,
	}, nil
}

func GetTrackInfo(filepath string) (TrackInfo, error) {
	probeResult, err := ProbeTrack(filepath)
	if err != nil {
		return TrackInfo{}, err
	}

	name := path.Base(filepath)
	trackName, err := ParseTrackName(name)
	if err != nil {
		return TrackInfo{}, err
	}

	return TrackInfo{
		Path:     filepath,
		Number:   trackName.Number,
		Name:     trackName.Name,
		Duration: probeResult.Duration,
		Tags:     probeResult.Tags,
	}, nil
}

func IsValidExt(exts []string, ext string) bool {
	if len(ext) == 0 {
		return false
	}

	if ext[0] == '.' {
		ext = ext[1:]
	}

	for _, valid := range exts {
		if valid == ext {
			return true
		}
	}

	return false
}

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
