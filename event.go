package emi_transport

import (
	milky_types "github.com/aK1r4z/emi-core/types"
)

func DefaultEventRegistries() map[milky_types.EventType]milky_types.Event {
	return map[milky_types.EventType]milky_types.Event{
		milky_types.EventBotOffline:                &milky_types.BotOfflineEvent{},
		milky_types.EventMessageReceive:            &milky_types.MessageReceiveEvent{},
		milky_types.EventMessageRecall:             &milky_types.MessageRecallEvent{},
		milky_types.EventFriendRequest:             &milky_types.FriendRequestEvent{},
		milky_types.EventGroupJoinRequest:          &milky_types.GroupJoinRequestEvent{},
		milky_types.EventGroupInvitedJoinRequest:   &milky_types.GroupInvitedJoinRequestEvent{},
		milky_types.EventGroupInvitation:           &milky_types.GroupInvitationEvent{},
		milky_types.EventFriendNudge:               &milky_types.FriendNudgeEvent{},
		milky_types.EventFriendFileUpload:          &milky_types.FriendFileUploadEvent{},
		milky_types.EventGroupAdminChange:          &milky_types.GroupAdminChangeEvent{},
		milky_types.EventGroupEssenceMessageChange: &milky_types.GroupEssenceMessageChangeEvent{},
		milky_types.EventGroupMemberIncrease:       &milky_types.GroupMemberIncreaseEvent{},
		milky_types.EventGroupMemberDecrease:       &milky_types.GroupMemberDecreaseEvent{},
		milky_types.EventGroupNameChange:           &milky_types.GroupNameChangeEvent{},
		milky_types.EventGroupMessageReaction:      &milky_types.GroupMessageReactionEvent{},
		milky_types.EventGroupMute:                 &milky_types.GroupMuteEvent{},
		milky_types.EventGroupWholeMute:            &milky_types.GroupWholeMuteEvent{},
		milky_types.EventGroupNudge:                &milky_types.GroupNudgeEvent{},
		milky_types.EventGroupFileUpload:           &milky_types.GroupFileUploadEvent{},
	}
}
