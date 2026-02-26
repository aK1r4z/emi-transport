package emi_transport

import (
	"context"

	emi_core "github.com/aK1r4z/emi-core"
)

type Logger interface {
	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)

	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
}

type EventSource interface {
	Open() (chan emi_core.RawEvent, error)
	Close() error
}

type APIClient interface {

	// SystemAPI

	GetLoginInfo(context.Context, emi_core.GetLoginInfoRequest) (*emi_core.GetLoginInfoResponse, error)                         // 获取登录信息
	GetImplInfo(context.Context, emi_core.GetImplInfoRequest) (*emi_core.GetImplInfoResponse, error)                            // 获取协议端信息
	GetUserProfile(context.Context, emi_core.GetUserProfileRequest) (*emi_core.GetUserProfileResponse, error)                   // 获取用户个人信息
	GetFriendList(context.Context, emi_core.GetFriendListRequest) (*emi_core.GetFriendListResponse, error)                      // 获取好友列表
	GetFriendInfo(context.Context, emi_core.GetFriendInfoRequest) (*emi_core.GetFriendInfoResponse, error)                      // 获取好友信息
	GetGroupList(context.Context, emi_core.GetGroupListRequest) (*emi_core.GetGroupListResponse, error)                         // 获取群列表
	GetGroupInfo(context.Context, emi_core.GetGroupInfoRequest) (*emi_core.GetGroupInfoResponse, error)                         // 获取群信息
	GetGroupMemberList(context.Context, emi_core.GetGroupMemberListRequest) (*emi_core.GetGroupMemberListResponse, error)       // 获取群成员列表
	GetGroupMemberInfo(context.Context, emi_core.GetGroupMemberInfoRequest) (*emi_core.GetGroupMemberInfoResponse, error)       // 获取群成员信息
	SetAvatar(context.Context, emi_core.SetAvatarRequest) (*emi_core.SetAvatarResponse, error)                                  // 设置 QQ 账号头像
	SetNickname(context.Context, emi_core.SetNicknameRequest) (*emi_core.SetNicknameResponse, error)                            // 设置 QQ 账号昵称
	SetBio(context.Context, emi_core.SetBioRequest) (*emi_core.SetBioResponse, error)                                           // 设置 QQ 账号个性签名
	GetCustomFaceURLList(context.Context, emi_core.GetCustomFaceURLListRequest) (*emi_core.GetCustomFaceURLListResponse, error) // 获取自定义表情 URL 列表
	GetCookies(context.Context, emi_core.GetCookiesRequest) (*emi_core.GetCookiesResponse, error)                               // 获取 Cookies
	GetCSRFToken(context.Context, emi_core.GetCSRFTokenRequest) (*emi_core.GetCSRFTokenResponse, error)                         // 获取 CSRF Token

	// MessageAPI

	SendPrivateMessage(context.Context, emi_core.SendPrivateMessageRequest) (*emi_core.SendPrivateMessageResponse, error)       // 发送私聊消息
	SendGroupMessage(context.Context, emi_core.SendGroupMessageRequest) (*emi_core.SendGroupMessageResponse, error)             // 发送群聊消息
	RecallPrivateMessage(context.Context, emi_core.RecallPrivateMessageRequest) (*emi_core.RecallPrivateMessageResponse, error) // 撤回私聊消息
	RecallGroupMessage(context.Context, emi_core.RecallGroupMessageRequest) (*emi_core.RecallGroupMessageResponse, error)       // 撤回群聊消息
	GetMessage(context.Context, emi_core.GetMessageRequest) (*emi_core.GetMessageResponse, error)                               // 获取消息
	GetHistoryMessages(context.Context, emi_core.GetHistoryMessagesRequest) (*emi_core.GetHistoryMessagesResponse, error)       // 获取历史消息列表
	GetResourceTempURL(context.Context, emi_core.GetResourceTempURLRequest) (*emi_core.GetResourceTempURLResponse, error)       // 获取临时资源链接
	GetForwardedMessages(context.Context, emi_core.GetForwardedMessagesRequest) (*emi_core.GetForwardedMessagesResponse, error) // 获取合并转发消息内容
	MarkMessageAsRead(context.Context, emi_core.MarkMessageAsReadRequest) (*emi_core.MarkMessageAsReadResponse, error)          // 标记消息为已读

	// FriendAPI

	SendFriendNudge(context.Context, emi_core.SendFriendNudgeRequest) (*emi_core.SendFriendNudgeResponse, error)             // 发送好友戳一戳
	SendProfileLike(context.Context, emi_core.SendProfileLikeRequest) (*emi_core.SendProfileLikeResponse, error)             // 发送名片点赞
	DeleteFriend(context.Context, emi_core.DeleteFriendRequest) (*emi_core.DeleteFriendResponse, error)                      // 删除好友
	GetFriendRequests(context.Context, emi_core.GetFriendRequestsRequest) (*emi_core.GetFriendRequestsResponse, error)       // 获取好友请求列表
	AcceptFriendRequest(context.Context, emi_core.AcceptFriendRequestRequest) (*emi_core.AcceptFriendRequestResponse, error) // 同意好友请求
	RejectFriendRequest(context.Context, emi_core.RejectFriendRequestRequest) (*emi_core.RejectFriendRequestResponse, error) // 拒绝好友请求

	// GroupAPI

	SetGroupName(context.Context, emi_core.SetGroupNameRequest) (*emi_core.SetGroupNameResponse, error)                                           // 设置群名称
	SetGroupAvatar(context.Context, emi_core.SetGroupAvatarRequest) (*emi_core.SetGroupAvatarResponse, error)                                     // 设置群头像
	SetGroupMemberCard(context.Context, emi_core.SetGroupMemberCardRequest) (*emi_core.SetGroupMemberCardResponse, error)                         // 设置群名片
	SetGroupMemberSpecialTitle(context.Context, emi_core.SetGroupMemberSpecialTitleRequest) (*emi_core.SetGroupMemberSpecialTitleResponse, error) // 设置群成员专属头衔
	SetGroupMemberAdmin(context.Context, emi_core.SetGroupMemberAdminRequest) (*emi_core.SetGroupMemberAdminResponse, error)                      // 设置群管理员
	SetGroupMemberMute(context.Context, emi_core.SetGroupMemberMuteRequest) (*emi_core.SetGroupMemberMuteResponse, error)                         // 设置群成员禁言
	SetGroupMemberWholeMute(context.Context, emi_core.SetGroupMemberWholeMuteRequest) (*emi_core.SetGroupMemberWholeMuteResponse, error)          // 设置群全员禁言
	KickGroupMember(context.Context, emi_core.KickGroupMemberRequest) (*emi_core.KickGroupMemberResponse, error)                                  // 踢出群成员
	GetGroupAnnouncements(context.Context, emi_core.GetGroupAnnouncementsRequest) (*emi_core.GetGroupAnnouncementsResponse, error)                // 获取群公告列表
	SendGroupAnnouncement(context.Context, emi_core.SendGroupAnnouncementRequest) (*emi_core.SendGroupAnnouncementResponse, error)                // 发送群公告
	DeleteGroupAnnouncement(context.Context, emi_core.DeleteGroupAnnouncementRequest) (*emi_core.DeleteGroupAnnouncementResponse, error)          // 删除群公告
	GetGroupEssenceMessages(context.Context, emi_core.GetGroupEssenceMessagesRequest) (*emi_core.GetGroupEssenceMessagesResponse, error)          // 获取群精华消息列表
	SetGroupEssenceMessage(context.Context, emi_core.SetGroupEssenceMessageRequest) (*emi_core.SetGroupEssenceMessageResponse, error)             // 设置群精华消息
	QuitGroup(context.Context, emi_core.QuitGroupRequest) (*emi_core.QuitGroupResponse, error)                                                    // 退出群
	SendGroupMessageReaction(context.Context, emi_core.SendGroupMessageReactionRequest) (*emi_core.SendGroupMessageReactionResponse, error)       // 发送群消息表情回应
	SendGroupNudge(context.Context, emi_core.SendGroupNudgeRequest) (*emi_core.SendGroupNudgeResponse, error)                                     // 发送群戳一戳
	GetGroupNotifications(context.Context, emi_core.GetGroupNotificationsRequest) (*emi_core.GetGroupNotificationsResponse, error)                // 获取群通知列表
	AcceptGroupRequest(context.Context, emi_core.AcceptGroupRequestRequest) (*emi_core.AcceptGroupRequestResponse, error)                         // 同意入群/邀请他人入群请求
	RejectGroupRequest(context.Context, emi_core.RejectGroupRequestRequest) (*emi_core.RejectGroupRequestResponse, error)                         // 拒绝入群/邀请他人入群请求
	AcceptGroupInvitation(context.Context, emi_core.AcceptGroupInvitationRequest) (*emi_core.AcceptGroupInvitationResponse, error)                // 同意他人邀请自身入群
	RejectGroupInvitation(context.Context, emi_core.RejectGroupInvitationRequest) (*emi_core.RejectGroupInvitationResponse, error)                // 拒绝他人邀请自身入群

	// FileAPI

	UploadPrivateFile(context.Context, emi_core.UploadPrivateFileRequest) (*emi_core.UploadPrivateFileResponse, error)                         // 上传私聊文件
	UploadGroupFile(context.Context, emi_core.UploadGroupFileRequest) (*emi_core.UploadGroupFileResponse, error)                               // 上传群文件
	GetPrivateFileDownloadURL(context.Context, emi_core.GetPrivateFileDownloadURLRequest) (*emi_core.GetPrivateFileDownloadURLResponse, error) // 获取私聊文件下载链接
	GetGroupFileDownloadURL(context.Context, emi_core.GetGroupFileDownloadURLRequest) (*emi_core.GetGroupFileDownloadURLResponse, error)       // 获取群文件下载链接
	GetGroupFiles(context.Context, emi_core.GetGroupFilesRequest) (*emi_core.GetGroupFilesResponse, error)                                     // 获取群文件列表
	MoveGroupFile(context.Context, emi_core.MoveGroupFileRequest) (*emi_core.MoveGroupFileResponse, error)                                     // 移动群文件
	RenameGroupFile(context.Context, emi_core.RenameGroupFileRequest) (*emi_core.RenameGroupFileResponse, error)                               // 重命名群文件
	DeleteGroupFile(context.Context, emi_core.DeleteGroupFileRequest) (*emi_core.DeleteGroupFileResponse, error)                               // 删除群文件
	CreateGroupFolder(context.Context, emi_core.CreateGroupFolderRequest) (*emi_core.CreateGroupFolderResponse, error)                         // 创建群文件夹
	RenameGroupFolder(context.Context, emi_core.RenameGroupFolderRequest) (*emi_core.RenameGroupFolderResponse, error)                         // 重命名群文件夹
	DeleteGroupFolder(context.Context, emi_core.DeleteGroupFolderRequest) (*emi_core.DeleteGroupFolderResponse, error)                         // 删除群文件夹
}
