package emi_transport

import (
	core "github.com/aK1r4z/emi-core"
	milky_types "github.com/aK1r4z/emi-core/types"
)

// [TODO] 检查是否需要将 core 包内的 segments 类型提到这个包中

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
	seg, err := core.NewTextSegment(text)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Mention(userID int64) *SegmentBuilder {
	seg, err := core.NewMentionSegment(userID)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) MentionAll() *SegmentBuilder {
	seg, err := core.NewMentionAllSegment()
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Face(faceID string, isLarge bool) *SegmentBuilder {
	seg, err := core.NewFaceSegment(faceID, isLarge)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Reply(messageSeq int64) *SegmentBuilder {
	seg, err := core.NewReplySegment(messageSeq)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Image(
	uri string,
	subType core.ImageSubtype,
	summary *string,
) *SegmentBuilder {
	seg, err := core.NewImageSegment(uri, subType, summary)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Record(uri string) *SegmentBuilder {
	seg, err := core.NewRecordSegment(uri)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Video(uri string, thumbURI *string) *SegmentBuilder {
	seg, err := core.NewVideoSegment(uri, thumbURI)
	return sb.Append(*seg, err)
}

func (sb *SegmentBuilder) Build() ([]milky_types.Segment, error) {
	if sb.error != nil {
		return nil, sb.error
	}
	return sb.segments, nil
}
