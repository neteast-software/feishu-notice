package feishunotice

import (
	"errors"
	"strings"

	"github.com/neteast-software/feishu-notice/internal/feishu"
)

const (
	defaultLocale  = "zh_cn"
	maxTitleLength = 120
)

// Message is a rich-text post message for Feishu.
type Message struct {
	Title      string
	Lines      []string
	Paragraphs []Paragraph
	Locale     Locale
}

// Paragraph is one paragraph in a Feishu post message.
type Paragraph []Segment

// SegmentTag is the Feishu post rich-text node type in the JSON tag field.
type SegmentTag string

const (
	// TagText renders plain text.
	TagText SegmentTag = "text"
	// TagLink renders a hyperlink.
	TagLink SegmentTag = "a"
	// TagAt mentions a user or all members.
	TagAt SegmentTag = "at"
	// TagImage renders an uploaded Feishu image.
	TagImage SegmentTag = "img"
)

// Segment is one Feishu post rich-text node.
type Segment struct {
	Tag      SegmentTag `json:"tag"`
	Text     string     `json:"text,omitempty"`
	Href     string     `json:"href,omitempty"`
	UserID   string     `json:"user_id,omitempty"`
	UserName string     `json:"user_name,omitempty"`
	ImageKey string     `json:"image_key,omitempty"`
	UnEscape *bool      `json:"un_escape,omitempty"`
}

// Text returns a text segment.
func Text(text string) Segment {
	return Segment{Tag: TagText, Text: text}
}

// Link returns a hyperlink segment.
func Link(text string, href string) Segment {
	return Segment{Tag: TagLink, Text: text, Href: href}
}

// At returns a mention segment. Use userID all to mention everyone.
func At(userID string, userName string) Segment {
	return Segment{Tag: TagAt, UserID: userID, UserName: userName}
}

// Image returns an image segment.
func Image(imageKey string) Segment {
	return Segment{Tag: TagImage, ImageKey: imageKey}
}

// Locale is a Feishu post locale key, for example zh_cn or en_us.
type Locale string

// String returns the locale value.
func (l Locale) String() string {
	if strings.TrimSpace(string(l)) == "" {
		return defaultLocale
	}
	return strings.TrimSpace(string(l))
}

// Validate checks whether the message can be sent.
func (m Message) Validate() error {
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("消息标题不能为空")
	}
	return nil
}

func (m Message) safeTitle() string {
	return truncate(m.Title, maxTitleLength)
}

func (m Message) feishuContent() [][]feishu.Segment {
	if len(m.Paragraphs) > 0 {
		content := make([][]feishu.Segment, 0, len(m.Paragraphs))
		for _, paragraph := range m.Paragraphs {
			content = append(content, paragraph.feishuSegments())
		}
		return content
	}
	content := make([][]feishu.Segment, 0, len(m.Lines))
	for _, line := range m.Lines {
		content = append(content, []feishu.Segment{{Tag: string(TagText), Text: line}})
	}
	return content
}

func (p Paragraph) feishuSegments() []feishu.Segment {
	segments := make([]feishu.Segment, 0, len(p))
	for _, segment := range p {
		segments = append(segments, segment.feishuSegment())
	}
	return segments
}

func (s Segment) feishuSegment() feishu.Segment {
	return feishu.Segment{
		Tag:      string(s.Tag),
		Text:     s.Text,
		Href:     s.Href,
		UserID:   s.UserID,
		UserName: s.UserName,
		ImageKey: s.ImageKey,
		UnEscape: s.UnEscape,
	}
}

func truncate(value string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxLength {
		return value
	}
	return string(runes[:maxLength])
}
