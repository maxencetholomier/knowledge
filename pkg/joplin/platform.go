package joplin

import "strings"

type JoplinPlatformDetector struct{}

func (d *JoplinPlatformDetector) DetectPlatform(url string) string {
	if strings.Contains(url, "localhost:41184") {
		return "Joplin"
	}
	return "serveur distant"
}
