package ffmpeg

import (
	"context"
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

	cmd := exec.CommandContext(ctx, binary+"probe", args...)
	output, err := cmd.Output()
	if err != nil {
		// ffprobe不可用时，用ffmpeg获取基本信息
		return probeWithFFmpeg(ctx, binary, inputPath)
	}

	_ = output
	// 简化解析，返回默认值
	return &ProbeResult{Duration: 0}, nil
}

func probeWithFFmpeg(ctx context.Context, binary, inputPath string) (*ProbeResult, error) {
	args := []string{"-i", inputPath}
	cmd := exec.CommandContext(ctx, binary, args...)
	output, _ := cmd.CombinedOutput()

	result := &ProbeResult{}
	text := string(output)

	// 解析 Duration
	if idx := strings.Index(text, "Duration:"); idx != -1 {
		durStr := text[idx+10 : idx+21]
		durStr = strings.TrimSpace(durStr)
		durStr = strings.TrimSuffix(durStr, ",")
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
