package apkinfo

import (
	"regexp"
	"strings"
)

type Info struct {
	Package     string
	VersionName string
	VersionCode string
	Label       string
	Launch      string
	MinSDK      string
	TargetSDK   string
}

var (
	rePackage  = regexp.MustCompile(`package:\s+name='([^']+)'`)
	reCode     = regexp.MustCompile(`versionCode='([^']*)'`)
	reName     = regexp.MustCompile(`versionName='([^']*)'`)
	reLaunch   = regexp.MustCompile(`launchable-activity:\s+name='([^']+)'`)
	reMinSDK   = regexp.MustCompile(`sdkVersion:'([^']+)'`)
	reTarget   = regexp.MustCompile(`targetSdkVersion:'([^']+)'`)
	reLabelDef = regexp.MustCompile(`application-label:'([^']*)'`)
	reLabelZH  = regexp.MustCompile(`application-label-zh(?:-[A-Za-z]+)?:'([^']*)'`)
)

func ParseBadging(text string) Info {
	info := Info{
		Package:     first(rePackage, text),
		VersionCode: first(reCode, text),
		VersionName: first(reName, text),
		Launch:      first(reLaunch, text),
		MinSDK:      first(reMinSDK, text),
		TargetSDK:   first(reTarget, text),
		Label:       first(reLabelZH, text),
	}
	if info.Label == "" {
		info.Label = first(reLabelDef, text)
	}
	return info
}

func first(re *regexp.Regexp, text string) string {
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}
