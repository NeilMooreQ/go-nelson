package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func CleanHTML(html string) string {
	noTags := strings.ReplaceAll(html, "<![CDATA[", "")
	noTags = strings.ReplaceAll(noTags, "]]>", "")

	noTags = strings.ReplaceAll(noTags, "<br>", "\n")
	noTags = strings.ReplaceAll(noTags, "<br />", "\n")

	noTags = RemoveHTMLTag(noTags, "img")

	noTags = RemoveAllHTMLTags(noTags)

	htmlEntities := map[string]string{
		"&mdash;": "—",
		"&laquo;": "«",
		"&raquo;": "»",
		"&nbsp;":  " ",
		"&times;": "×",
		"&ndash;": "–",
		"&quot;":  "\"",
		"&#039;":  "'",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
	}

	for entity, replacement := range htmlEntities {
		noTags = strings.ReplaceAll(noTags, entity, replacement)
	}

	noTags = strings.TrimSpace(noTags)
	noTags = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(noTags, "\n\n")

	return noTags
}

func RemoveHTMLTag(s string, tag string) string {
	startTag := "<" + tag
	endTag := "</" + tag + ">"

	for {
		startIdx := strings.Index(s, startTag)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(s[startIdx:], ">")
		if endIdx == -1 {
			break
		}

		endIdx += startIdx + 1

		if strings.Contains(s[startIdx:endIdx], "/>") {
			s = s[:startIdx] + s[endIdx:]
			continue
		}

		closeTagIdx := strings.Index(s[endIdx:], endTag)
		if closeTagIdx == -1 {
			break
		}

		closeTagIdx += endIdx + len(endTag)

		s = s[:startIdx] + s[closeTagIdx:]
	}

	return s
}

func RemoveAllHTMLTags(s string) string {
	var result string
	var inTag bool

	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result += string(r)
		}
	}

	return result
}

func ExtractImageURL(htmlContent string) string {
	imgRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	matches := imgRegex.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func CleanCDATA(text string) string {
	text = strings.ReplaceAll(text, "<![CDATA[", "")
	text = strings.ReplaceAll(text, "]]>", "")
	return text
}

func ParseRSSDate(dateStr string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 +0300",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("невозможно распарсить дату: %s", dateStr)
}

func ReplaceHTMLEntity(html, entity, replacement string) string {
	return strings.ReplaceAll(html, entity, replacement)
}

func FindNthOccurrence(s, substr string, n int) int {
	count := 0
	idx := 0

	for {
		found := strings.Index(s[idx:], substr)
		if found == -1 {
			return -1
		}

		idx += found
		count++

		if count == n {
			return idx
		}

		idx += len(substr)

		if idx >= len(s) {
			return -1
		}
	}
}
