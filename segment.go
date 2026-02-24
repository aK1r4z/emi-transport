package emi_transport

import (
	"encoding/json"

	milky_types "github.com/aK1r4z/emi-core/types"
)

type ImageSubtype string

const (
	ImageNormal  ImageSubtype = "normal"
	ImageSticker ImageSubtype = "sticker"
)

type Element interface {
	Type() milky_types.SegmentType
	json.Marshaler
	json.Unmarshaler
}

func NewSegment(segmentType milky_types.SegmentType, element Element) (*milky_types.Segment, error) {
	b, err := element.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &milky_types.Segment{
		Type: segmentType,
		Data: json.RawMessage(b),
	}, nil
}

// 文本消息段
type TextElement struct {
	Text string `json:"text"` // 文本内容
}

func (s *TextElement) Type() milky_types.SegmentType { return milky_types.SegmentText }
func (s *TextElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *TextElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewTextSegment(text string) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentText, &TextElement{Text: text})
}

// 提及消息段
type MentionElement struct {
	UserID int64 `json:"user_id"` // 提及的 QQ 号
}

func (s *MentionElement) Type() milky_types.SegmentType { return milky_types.SegmentMention }
func (s *MentionElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *MentionElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewMentionSegment(userID int64) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentMention, &MentionElement{UserID: userID})
}

// 提及全体消息段
type MentionAllElement struct{}

func (s *MentionAllElement) Type() milky_types.SegmentType { return milky_types.SegmentMentionAll }
func (s *MentionAllElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *MentionAllElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewMentionAllSegment() (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentMentionAll, &MentionAllElement{})
}

// 表情消息段
type FaceElement struct {
	FaceID  string `json:"face_id"`  // 表情 ID
	IsLarge bool   `json:"is_large"` // 是否为超级表情
}

func (s *FaceElement) Type() milky_types.SegmentType { return milky_types.SegmentFace }
func (s *FaceElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *FaceElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewFaceSegment(faceID string, isLarge bool) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentFace, &FaceElement{FaceID: faceID, IsLarge: isLarge})
}

// 回复消息段
type ReplyElement struct {
	MessageSeq int64 `json:"message_seq"` // 被引用的消息序列号
}

func (s *ReplyElement) Type() milky_types.SegmentType { return milky_types.SegmentReply }
func (s *ReplyElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *ReplyElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewReplySegment(messageSeq int64) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentReply, &ReplyElement{MessageSeq: messageSeq})
}

// 图片消息段
type ImageElement struct {
	URI     string       `json:"uri"`      // 文件 URI，支持 file:// http(s):// base64:// 三种格式
	SubType ImageSubtype `json:"sub_type"` // 图片类型，可能值：normal sticker，默认值：normal
	Summary *string      `json:"summary"`  // 图片预览文本
}

func (s *ImageElement) Type() milky_types.SegmentType { return milky_types.SegmentImage }
func (s *ImageElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *ImageElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewImageSegment(
	uri string,
	subType ImageSubtype,
	summary *string,
) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentImage, &ImageElement{
		URI:     uri,
		SubType: subType,
		Summary: summary,
	})
}

// 语音消息段
type RecordElement struct {
	URI string `json:"uri"` // 文件 URI，支持 file:// http(s):// base64:// 三种格式
}

func (s *RecordElement) Type() milky_types.SegmentType { return milky_types.SegmentRecord }
func (s *RecordElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *RecordElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewRecordSegment(uri string) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentRecord, &RecordElement{URI: uri})
}

// 视频消息段
type VideoElement struct {
	URI      string  `json:"uri"`       // 文件 URI，支持 file:// http(s):// base64:// 三种格式
	ThumbURI *string `json:"thumb_uri"` // 封面图片 URI
}

func (s *VideoElement) Type() milky_types.SegmentType { return milky_types.SegmentVideo }
func (s *VideoElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *VideoElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewVideoSegment(uri string, thumbURI *string) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentVideo, &VideoElement{URI: uri, ThumbURI: thumbURI})
}

// 合并转发消息段
type ForwardElement struct {
	Messages []milky_types.OutgoingForwardedMessage `json:"messages"` // 合并转发消息段
}

func (s *ForwardElement) Type() milky_types.SegmentType { return milky_types.SegmentForward }
func (s *ForwardElement) MarshalJSON() ([]byte, error)  { return json.Marshal(*s) }
func (s *ForwardElement) UnmarshalJSON(b []byte) error  { return json.Unmarshal(b, s) }

func NewForwardSegment(messages []milky_types.OutgoingForwardedMessage) (*milky_types.Segment, error) {
	return NewSegment(milky_types.SegmentForward, &ForwardElement{Messages: messages})
}

type SegmentBuilder struct {
	segments []milky_types.Segment
	error    error
}

func NewSegmentBuilder() *SegmentBuilder {
	return &SegmentBuilder{
		segments: make([]milky_types.Segment, 0),
	}
}

func (sb *SegmentBuilder) Append(segment milky_types.Segment, err error) *SegmentBuilder {
	if sb.error != nil {
		return sb
	}
	if err != nil {
		sb.error = err
		return sb
	}
	sb.segments = append(sb.segments, segment)
	return sb
}

func (sb *SegmentBuilder) Text(text string) *SegmentBuilder {
	seg, err := NewTextSegment(text)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Mention(userID int64) *SegmentBuilder {
	seg, err := NewMentionSegment(userID)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) MentionAll() *SegmentBuilder {
	seg, err := NewMentionAllSegment()
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Face(faceID string, isLarge bool) *SegmentBuilder {
	seg, err := NewFaceSegment(faceID, isLarge)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Reply(messageSeq int64) *SegmentBuilder {
	seg, err := NewReplySegment(messageSeq)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Image(
	uri string,
	subType ImageSubtype,
	summary *string,
) *SegmentBuilder {
	seg, err := NewImageSegment(uri, subType, summary)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Record(uri string) *SegmentBuilder {
	seg, err := NewRecordSegment(uri)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Video(uri string, thumbURI *string) *SegmentBuilder {
	seg, err := NewVideoSegment(uri, thumbURI)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Build() ([]milky_types.Segment, error) {
	if sb.error != nil {
		return nil, sb.error
	}
	return sb.segments, nil
}
