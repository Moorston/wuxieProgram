package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type ProbeResult struct {
	Duration float64
	Width    int
	Height   int
	Bitrate  int
}

func Probe(ctx context.Context, binary, inputPath string) (*ProbeResult, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		// ffprobe不可用时，用ffmpeg获取基本信息
		return probeWithFFmpeg(ctx, binary, inputPath)
	}

	var probeData struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probeData); err != nil {
		return probeWithFFmpeg(ctx, binary, inputPath)
	}

	result := &ProbeResult{}
	if dur, err := strconv.ParseFloat(probeData.Format.Duration, 64); err == nil {
		result.Duration = dur
	}
	if br, err := strconv.Atoi(probeData.Format.BitRate); err == nil {
		result.Bitrate = br
	}

	for _, stream := range probeData.Streams {
		if stream.CodecType == "video" {
			result.Width = stream.Width
			result.Height = stream.Height
			break
		}
	}

	return result, nil
}

func probeWithFFmpeg(ctx context.Context, binary, inputPath string) (*ProbeResult, error) {
	args := []string{"-i", inputPath}
	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg probe failed: %w", err)
	}

	result := &ProbeResult{}
	text := string(output)

	// 解析 Duration - 更健壮的解析逻辑
	if idx := strings.Index(text, "Duration:"); idx != -1 {
		durStart := idx + 10
		// 找到Duration行的结尾（通常是逗号或换行）
		durEnd := strings.IndexAny(text[durStart:], ",\n")
		if durEnd == -1 {
			durEnd = len(text)
		} else {
			durEnd += durStart
		}

		durStr := strings.TrimSpace(text[durStart:durEnd])
		parts := strings.Split(durStr, ":")
		if len(parts) == 3 {
			h, _ := strconv.ParseFloat(parts[0], 64)
			m, _ := strconv.ParseFloat(parts[1], 64)
			s, _ := strconv.ParseFloat(parts[2], 64)
			result.Duration = h*3600 + m*60 + s
		}
	}

	return result, nil
}

type TranscodeOptions struct {
	InputPath  string
	OutputPath string
	CoverPath  string
	MaxWidth   int
	MaxBitrate string
}

func Transcode(ctx context.Context, binary string, opts TranscodeOptions) (float64, error) {
	if opts.MaxWidth == 0 {
		opts.MaxWidth = 1280
	}
	if opts.MaxBitrate == "" {
		opts.MaxBitrate = "2000k"
	}

	args := []string{
		"-i", opts.InputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "28",
		"-vf", fmt.Sprintf("scale='min(%d,iw)':-2", opts.MaxWidth),
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		opts.OutputPath,
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("transcode failed: %w, output: %s", err, string(output))
	}

	// 提取封面
	if opts.CoverPath != "" {
		coverArgs := []string{
			"-i", opts.OutputPath,
			"-ss", "1",
			"-frames:v", "1",
			"-q:v", "2",
			"-y",
			opts.CoverPath,
		}
		exec.CommandContext(ctx, binary, coverArgs...).Run()
	}

	// 获取转码后时长
	probe, _ := Probe(ctx, binary, opts.OutputPath)
	if probe != nil {
		return probe.Duration, nil
	}

	return 0, nil
}
