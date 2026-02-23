package emi_transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	core "github.com/aK1r4z/emi-core"
	"github.com/aK1r4z/emi-core/api"
)

type HttpResult struct {
	Status string          `json:"status"`
	Code   int             `json:"retcode"`
	Data   json.RawMessage `json:"data"`
}

type HttpClient struct {
	logger core.Logger

	restGateway string
	accessToken string

	client http.Client

	maxRetries int

	baseRetryDelay time.Duration
	maxRetryDelay  time.Duration
	maxRetryJitter time.Duration
}

func NewHttpClient(logger core.Logger, restGateway string, accessToken string) *HttpClient {
	return &HttpClient{
		logger: logger,

		restGateway: restGateway,
		accessToken: accessToken,

		client: http.Client{
			Timeout: time.Second * 10,
		},

		maxRetries: 5,

		baseRetryDelay: 100 * time.Millisecond,
		maxRetryDelay:  5 * time.Second,
		maxRetryJitter: 100 * time.Millisecond,
	}
}

func NewHttpClientWithOptions(
	logger core.Logger,

	restGateway string,
	accessToken string,

	client http.Client,

	maxRetries int,

	baseRetryDelay time.Duration,
	maxRetryDelay time.Duration,
	maxRetryJitter time.Duration,
) *HttpClient {
	return &HttpClient{
		logger: logger,

		restGateway: restGateway,
		accessToken: accessToken,

		client: client,

		maxRetries: maxRetries,

		baseRetryDelay: baseRetryDelay,
		maxRetryDelay:  maxRetryDelay,
		maxRetryJitter: maxRetryJitter,
	}
}

func (h *HttpClient) Post(ctx context.Context, endpoint string, request any, response any) error {
	h.logger.Debugf("Sending post request to %s", endpoint)
	urlPath, err := url.JoinPath(h.restGateway, endpoint)
	if err != nil {
		return fmt.Errorf("failed to join URL path: %w", err)
	}
	h.logger.Debugf("URL path: %s", urlPath)

	attempt := 0

	for {
		err := h.doPost(ctx, urlPath, request, response)
		if err == nil {
			return nil
		} else if attempt > h.maxRetries {
			return fmt.Errorf("max retries exceeded: %w", err)
		}

		// 请求失败，开始重试
		jitter := time.Duration(rand.Int64N(int64(h.maxRetryJitter)))
		delay := min(
			h.baseRetryDelay*(1<<attempt)+jitter,
			h.maxRetryDelay,
		)

		h.logger.Debugf("Retrying request to %s after %s (attempt %d/%d)", endpoint, delay, attempt, h.maxRetries)

		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled")
		case <-time.After(delay):
		}

		attempt += 1
	}
}

func (h *HttpClient) doPost(ctx context.Context, urlPath string, request any, response any) error {

	// 构建 HTTP 请求体
	var bodyReader io.Reader = bytes.NewReader([]byte{})
	if request != nil {
		jsonBytes, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		h.logger.Debugf("Request body: %s", string(jsonBytes))
		bodyReader = bytes.NewReader(jsonBytes)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlPath, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if h.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.accessToken)
	}

	// 发送 HTTP 请求
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取请求结果
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	h.logger.Debugf("response body: %s", string(body))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status code %d, response body: %s", resp.StatusCode, string(body))
	}

	if response == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// 解码请求结果
	result := HttpResult{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if err := json.Unmarshal(result.Data, response); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// SystemAPI

// 获取登录信息
func (h *HttpClient) GetLoginInfo(ctx context.Context, request api.GetLoginInfoRequest) (*api.GetLoginInfoResponse, error) {
	var resp api.GetLoginInfoResponse
	if err := h.Post(ctx, string(core.GetLoginInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取协议端信息
func (h *HttpClient) GetImplInfo(ctx context.Context, request api.GetImplInfoRequest) (*api.GetImplInfoResponse, error) {
	var resp api.GetImplInfoResponse
	if err := h.Post(ctx, string(core.GetImplInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取用户个人信息
func (h *HttpClient) GetUserProfile(ctx context.Context, request api.GetUserProfileRequest) (*api.GetUserProfileResponse, error) {
	var resp api.GetUserProfileResponse
	if err := h.Post(ctx, string(core.GetUserProfile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取好友列表
func (h *HttpClient) GetFriendList(ctx context.Context, request api.GetFriendListRequest) (*api.GetFriendListResponse, error) {
	var resp api.GetFriendListResponse
	if err := h.Post(ctx, string(core.GetFriendList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取好友信息
func (h *HttpClient) GetFriendInfo(ctx context.Context, request api.GetFriendInfoRequest) (*api.GetFriendInfoResponse, error) {
	var resp api.GetFriendInfoResponse
	if err := h.Post(ctx, string(core.GetFriendInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群列表
func (h *HttpClient) GetGroupList(ctx context.Context, request api.GetGroupListRequest) (*api.GetGroupListResponse, error) {
	var resp api.GetGroupListResponse
	if err := h.Post(ctx, string(core.GetGroupList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群信息
func (h *HttpClient) GetGroupInfo(ctx context.Context, request api.GetGroupInfoRequest) (*api.GetGroupInfoResponse, error) {
	var resp api.GetGroupInfoResponse
	if err := h.Post(ctx, string(core.GetGroupInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群成员列表
func (h *HttpClient) GetGroupMemberList(ctx context.Context, request api.GetGroupMemberListRequest) (*api.GetGroupMemberListResponse, error) {
	var resp api.GetGroupMemberListResponse
	if err := h.Post(ctx, string(core.GetGroupMemberList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群成员信息
func (h *HttpClient) GetGroupMemberInfo(ctx context.Context, request api.GetGroupMemberInfoRequest) (*api.GetGroupMemberInfoResponse, error) {
	var resp api.GetGroupMemberInfoResponse
	if err := h.Post(ctx, string(core.GetGroupMemberInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置 QQ 账号头像
func (h *HttpClient) SetAvatar(ctx context.Context, request api.SetAvatarRequest) (*api.SetAvatarResponse, error) {
	var resp api.SetAvatarResponse
	if err := h.Post(ctx, string(core.SetAvatar), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置 QQ 账号昵称
func (h *HttpClient) SetNickname(ctx context.Context, request api.SetNicknameRequest) (*api.SetNicknameResponse, error) {
	var resp api.SetNicknameResponse
	if err := h.Post(ctx, string(core.SetNickname), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置 QQ 账号个性签名
func (h *HttpClient) SetBio(ctx context.Context, request api.SetBioRequest) (*api.SetBioResponse, error) {
	var resp api.SetBioResponse
	if err := h.Post(ctx, string(core.SetBio), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取自定义表情 URL 列表
func (h *HttpClient) GetCustomFaceURLList(ctx context.Context, request api.GetCustomFaceURLListRequest) (*api.GetCustomFaceURLListResponse, error) {
	var resp api.GetCustomFaceURLListResponse
	if err := h.Post(ctx, string(core.GetCustomFaceURLList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取 Cookies
func (h *HttpClient) GetCookies(ctx context.Context, request api.GetCookiesRequest) (*api.GetCookiesResponse, error) {
	var resp api.GetCookiesResponse
	if err := h.Post(ctx, string(core.GetCookies), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取 CSRF Token
func (h *HttpClient) GetCSRFToken(ctx context.Context, request api.GetCSRFTokenRequest) (*api.GetCSRFTokenResponse, error) {
	var resp api.GetCSRFTokenResponse
	if err := h.Post(ctx, string(core.GetCSRFToken), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MessageAPI

// 发送私聊消息
func (h *HttpClient) SendPrivateMessage(ctx context.Context, request api.SendPrivateMessageRequest) (*api.SendPrivateMessageResponse, error) {
	var resp api.SendPrivateMessageResponse
	if err := h.Post(ctx, string(core.SendPrivateMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群聊消息
func (h *HttpClient) SendGroupMessage(ctx context.Context, request api.SendGroupMessageRequest) (*api.SendGroupMessageResponse, error) {
	var resp api.SendGroupMessageResponse
	if err := h.Post(ctx, string(core.SendGroupMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 撤回私聊消息
func (h *HttpClient) RecallPrivateMessage(ctx context.Context, request api.RecallPrivateMessageRequest) (*api.RecallPrivateMessageResponse, error) {
	var resp api.RecallPrivateMessageResponse
	if err := h.Post(ctx, string(core.RecallPrivateMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 撤回群聊消息
func (h *HttpClient) RecallGroupMessage(ctx context.Context, request api.RecallGroupMessageRequest) (*api.RecallGroupMessageResponse, error) {
	var resp api.RecallGroupMessageResponse
	if err := h.Post(ctx, string(core.RecallGroupMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取消息
func (h *HttpClient) GetMessage(ctx context.Context, request api.GetMessageRequest) (*api.GetMessageResponse, error) {
	var resp api.GetMessageResponse
	if err := h.Post(ctx, string(core.GetMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取历史消息列表
func (h *HttpClient) GetHistoryMessages(ctx context.Context, request api.GetHistoryMessagesRequest) (*api.GetHistoryMessagesResponse, error) {
	var resp api.GetHistoryMessagesResponse
	if err := h.Post(ctx, string(core.GetHistoryMessages), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取临时资源链接
func (h *HttpClient) GetResourceTempURL(ctx context.Context, request api.GetResourceTempURLRequest) (*api.GetResourceTempURLResponse, error) {
	var resp api.GetResourceTempURLResponse
	if err := h.Post(ctx, string(core.GetResourceTempURL), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取合并转发消息内容
func (h *HttpClient) GetForwardedMessages(ctx context.Context, request api.GetForwardedMessagesRequest) (*api.GetForwardedMessagesResponse, error) {
	var resp api.GetForwardedMessagesResponse
	if err := h.Post(ctx, string(core.GetForwardedMessages), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 标记消息为已读
func (h *HttpClient) MarkMessageAsRead(ctx context.Context, request api.MarkMessageAsReadRequest) (*api.MarkMessageAsReadResponse, error) {
	var resp api.MarkMessageAsReadResponse
	if err := h.Post(ctx, string(core.MarkMessageAsRead), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FriendAPI

// 发送好友戳一戳
func (h *HttpClient) SendFriendNudge(ctx context.Context, request api.SendFriendNudgeRequest) (*api.SendFriendNudgeResponse, error) {
	var resp api.SendFriendNudgeResponse
	if err := h.Post(ctx, string(core.SendFriendNudge), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送名片点赞
func (h *HttpClient) SendProfileLike(ctx context.Context, request api.SendProfileLikeRequest) (*api.SendProfileLikeResponse, error) {
	var resp api.SendProfileLikeResponse
	if err := h.Post(ctx, string(core.SendProfileLike), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除好友
func (h *HttpClient) DeleteFriend(ctx context.Context, request api.DeleteFriendRequest) (*api.DeleteFriendResponse, error) {
	var resp api.DeleteFriendResponse
	if err := h.Post(ctx, string(core.DeleteFriend), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取好友请求列表
func (h *HttpClient) GetFriendRequests(ctx context.Context, request api.GetFriendRequestsRequest) (*api.GetFriendRequestsResponse, error) {
	var resp api.GetFriendRequestsResponse
	if err := h.Post(ctx, string(core.GetFriendRequests), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 同意好友请求
func (h *HttpClient) AcceptFriendRequest(ctx context.Context, request api.AcceptFriendRequestRequest) (*api.AcceptFriendRequestResponse, error) {
	var resp api.AcceptFriendRequestResponse
	if err := h.Post(ctx, string(core.AcceptFriendRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 拒绝好友请求
func (h *HttpClient) RejectFriendRequest(ctx context.Context, request api.RejectFriendRequestRequest) (*api.RejectFriendRequestResponse, error) {
	var resp api.RejectFriendRequestResponse
	if err := h.Post(ctx, string(core.RejectFriendRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GroupAPI

// 设置群名称
func (h *HttpClient) SetGroupName(ctx context.Context, request api.SetGroupNameRequest) (*api.SetGroupNameResponse, error) {
	var resp api.SetGroupNameResponse
	if err := h.Post(ctx, string(core.SetGroupName), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群头像
func (h *HttpClient) SetGroupAvatar(ctx context.Context, request api.SetGroupAvatarRequest) (*api.SetGroupAvatarResponse, error) {
	var resp api.SetGroupAvatarResponse
	if err := h.Post(ctx, string(core.SetGroupAvatar), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群名片
func (h *HttpClient) SetGroupMemberCard(ctx context.Context, request api.SetGroupMemberCardRequest) (*api.SetGroupMemberCardResponse, error) {
	var resp api.SetGroupMemberCardResponse
	if err := h.Post(ctx, string(core.SetGroupMemberCard), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群成员专属头衔
func (h *HttpClient) SetGroupMemberSpecialTitle(ctx context.Context, request api.SetGroupMemberSpecialTitleRequest) (*api.SetGroupMemberSpecialTitleResponse, error) {
	var resp api.SetGroupMemberSpecialTitleResponse
	if err := h.Post(ctx, string(core.SetGroupMemberSpecialTitle), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群管理员
func (h *HttpClient) SetGroupMemberAdmin(ctx context.Context, request api.SetGroupMemberAdminRequest) (*api.SetGroupMemberAdminResponse, error) {
	var resp api.SetGroupMemberAdminResponse
	if err := h.Post(ctx, string(core.SetGroupMemberAdmin), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群成员禁言
func (h *HttpClient) SetGroupMemberMute(ctx context.Context, request api.SetGroupMemberMuteRequest) (*api.SetGroupMemberMuteResponse, error) {
	var resp api.SetGroupMemberMuteResponse
	if err := h.Post(ctx, string(core.SetGroupMemberMute), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群全员禁言
func (h *HttpClient) SetGroupMemberWholeMute(ctx context.Context, request api.SetGroupMemberWholeMuteRequest) (*api.SetGroupMemberWholeMuteResponse, error) {
	var resp api.SetGroupMemberWholeMuteResponse
	if err := h.Post(ctx, string(core.SetGroupMemberWholeMute), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 踢出群成员
func (h *HttpClient) KickGroupMember(ctx context.Context, request api.KickGroupMemberRequest) (*api.KickGroupMemberResponse, error) {
	var resp api.KickGroupMemberResponse
	if err := h.Post(ctx, string(core.KickGroupMember), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群公告列表
func (h *HttpClient) GetGroupAnnouncements(ctx context.Context, request api.GetGroupAnnouncementsRequest) (*api.GetGroupAnnouncementsResponse, error) {
	var resp api.GetGroupAnnouncementsResponse
	if err := h.Post(ctx, string(core.GetGroupAnnouncements), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群公告
func (h *HttpClient) SendGroupAnnouncement(ctx context.Context, request api.SendGroupAnnouncementRequest) (*api.SendGroupAnnouncementResponse, error) {
	var resp api.SendGroupAnnouncementResponse
	if err := h.Post(ctx, string(core.SendGroupAnnouncement), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除群公告
func (h *HttpClient) DeleteGroupAnnouncement(ctx context.Context, request api.DeleteGroupAnnouncementRequest) (*api.DeleteGroupAnnouncementResponse, error) {
	var resp api.DeleteGroupAnnouncementResponse
	if err := h.Post(ctx, string(core.DeleteGroupAnnouncement), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群精华消息列表
func (h *HttpClient) GetGroupEssenceMessages(ctx context.Context, request api.GetGroupEssenceMessagesRequest) (*api.GetGroupEssenceMessagesResponse, error) {
	var resp api.GetGroupEssenceMessagesResponse
	if err := h.Post(ctx, string(core.GetGroupEssenceMessages), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群精华消息
func (h *HttpClient) SetGroupEssenceMessage(ctx context.Context, request api.SetGroupEssenceMessageRequest) (*api.SetGroupEssenceMessageResponse, error) {
	var resp api.SetGroupEssenceMessageResponse
	if err := h.Post(ctx, string(core.SetGroupEssenceMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 退出群
func (h *HttpClient) QuitGroup(ctx context.Context, request api.QuitGroupRequest) (*api.QuitGroupResponse, error) {
	var resp api.QuitGroupResponse
	if err := h.Post(ctx, string(core.QuitGroup), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群消息表情回应
func (h *HttpClient) SendGroupMessageReaction(ctx context.Context, request api.SendGroupMessageReactionRequest) (*api.SendGroupMessageReactionResponse, error) {
	var resp api.SendGroupMessageReactionResponse
	if err := h.Post(ctx, string(core.SendGroupMessageReaction), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群戳一戳
func (h *HttpClient) SendGroupNudge(ctx context.Context, request api.SendGroupNudgeRequest) (*api.SendGroupNudgeResponse, error) {
	var resp api.SendGroupNudgeResponse
	if err := h.Post(ctx, string(core.SendGroupNudge), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群通知列表
func (h *HttpClient) GetGroupNotifications(ctx context.Context, request api.GetGroupNotificationsRequest) (*api.GetGroupNotificationsResponse, error) {
	var resp api.GetGroupNotificationsResponse
	if err := h.Post(ctx, string(core.GetGroupNotifications), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 同意入群/邀请他人入群请求
func (h *HttpClient) AcceptGroupRequest(ctx context.Context, request api.AcceptGroupRequestRequest) (*api.AcceptGroupRequestResponse, error) {
	var resp api.AcceptGroupRequestResponse
	if err := h.Post(ctx, string(core.AcceptGroupRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 拒绝入群/邀请他人入群请求
func (h *HttpClient) RejectGroupRequest(ctx context.Context, request api.RejectGroupRequestRequest) (*api.RejectGroupRequestResponse, error) {
	var resp api.RejectGroupRequestResponse
	if err := h.Post(ctx, string(core.RejectGroupRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 同意他人邀请自身入群
func (h *HttpClient) AcceptGroupInvitation(ctx context.Context, request api.AcceptGroupInvitationRequest) (*api.AcceptGroupInvitationResponse, error) {
	var resp api.AcceptGroupInvitationResponse
	if err := h.Post(ctx, string(core.AcceptGroupInvitation), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 拒绝他人邀请自身入群
func (h *HttpClient) RejectGroupInvitation(ctx context.Context, request api.RejectGroupInvitationRequest) (*api.RejectGroupInvitationResponse, error) {
	var resp api.RejectGroupInvitationResponse
	if err := h.Post(ctx, string(core.RejectGroupInvitation), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FileAPI

// 上传私聊文件
func (h *HttpClient) UploadPrivateFile(ctx context.Context, request api.UploadPrivateFileRequest) (*api.UploadPrivateFileResponse, error) {
	var resp api.UploadPrivateFileResponse
	if err := h.Post(ctx, string(core.UploadPrivateFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 上传群文件
func (h *HttpClient) UploadGroupFile(ctx context.Context, request api.UploadGroupFileRequest) (*api.UploadGroupFileResponse, error) {
	var resp api.UploadGroupFileResponse
	if err := h.Post(ctx, string(core.UploadGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取私聊文件下载链接
func (h *HttpClient) GetPrivateFileDownloadURL(ctx context.Context, request api.GetPrivateFileDownloadURLRequest) (*api.GetPrivateFileDownloadURLResponse, error) {
	var resp api.GetPrivateFileDownloadURLResponse
	if err := h.Post(ctx, string(core.GetPrivateFileDownloadURL), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群文件下载链接
func (h *HttpClient) GetGroupFileDownloadURL(ctx context.Context, request api.GetGroupFileDownloadURLRequest) (*api.GetGroupFileDownloadURLResponse, error) {
	var resp api.GetGroupFileDownloadURLResponse
	if err := h.Post(ctx, string(core.GetGroupFileDownloadURL), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群文件列表
func (h *HttpClient) GetGroupFiles(ctx context.Context, request api.GetGroupFilesRequest) (*api.GetGroupFilesResponse, error) {
	var resp api.GetGroupFilesResponse
	if err := h.Post(ctx, string(core.GetGroupFiles), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 移动群文件
func (h *HttpClient) MoveGroupFile(ctx context.Context, request api.MoveGroupFileRequest) (*api.MoveGroupFileResponse, error) {
	var resp api.MoveGroupFileResponse
	if err := h.Post(ctx, string(core.MoveGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 重命名群文件
func (h *HttpClient) RenameGroupFile(ctx context.Context, request api.RenameGroupFileRequest) (*api.RenameGroupFileResponse, error) {
	var resp api.RenameGroupFileResponse
	if err := h.Post(ctx, string(core.RenameGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除群文件
func (h *HttpClient) DeleteGroupFile(ctx context.Context, request api.DeleteGroupFileRequest) (*api.DeleteGroupFileResponse, error) {
	var resp api.DeleteGroupFileResponse
	if err := h.Post(ctx, string(core.DeleteGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 创建群文件夹
func (h *HttpClient) CreateGroupFolder(ctx context.Context, request api.CreateGroupFolderRequest) (*api.CreateGroupFolderResponse, error) {
	var resp api.CreateGroupFolderResponse
	if err := h.Post(ctx, string(core.CreateGroupFolder), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 重命名群文件夹
func (h *HttpClient) RenameGroupFolder(ctx context.Context, request api.RenameGroupFolderRequest) (*api.RenameGroupFolderResponse, error) {
	var resp api.RenameGroupFolderResponse
	if err := h.Post(ctx, string(core.RenameGroupFolder), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除群文件夹
func (h *HttpClient) DeleteGroupFolder(ctx context.Context, request api.DeleteGroupFolderRequest) (*api.DeleteGroupFolderResponse, error) {
	var resp api.DeleteGroupFolderResponse
	if err := h.Post(ctx, string(core.DeleteGroupFolder), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
