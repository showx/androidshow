package apkinfo

import "testing"

func TestParseBadging(t *testing.T) {
	raw := `package: name='com.example.app' versionCode='12' versionName='1.2.0' compileSdkVersion='34'
sdkVersion:'21'
targetSdkVersion:'34'
application-label:'Example'
application-label-zh:'示例应用'
application-label-zh-CN:'示例应用'
launchable-activity: name='com.example.app.MainActivity'  label='Example' icon=''
`
	info := ParseBadging(raw)
	if info.Package != "com.example.app" {
		t.Fatalf("package = %q", info.Package)
	}
	if info.VersionCode != "12" || info.VersionName != "1.2.0" {
		t.Fatalf("version = %s (%s)", info.VersionName, info.VersionCode)
	}
	if info.Label != "示例应用" {
		t.Fatalf("label = %q", info.Label)
	}
	if info.Launch != "com.example.app.MainActivity" {
		t.Fatalf("launch = %q", info.Launch)
	}
	if info.MinSDK != "21" || info.TargetSDK != "34" {
		t.Fatalf("sdk = %s/%s", info.MinSDK, info.TargetSDK)
	}
}
