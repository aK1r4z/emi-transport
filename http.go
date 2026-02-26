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

	emi_core "github.com/aK1r4z/emi-core"
)

type HttpResult struct {
	Status string          `json:"status"`
	Code   int             `json:"retcode"`
	Data   json.RawMessage `json:"data"`
}

type HttpClient struct {
	logger Logger

	restGateway string
	accessToken string

	client http.Client

	maxRetries int

	baseRetryDelay time.Duration
	maxRetryDelay  time.Duration
	maxRetryJitter time.Duration
}

func NewHttpClient(logger Logger, restGateway string, accessToken string) *HttpClient {
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
	logger Logger,

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
func (h *HttpClient) GetLoginInfo(ctx context.Context, request emi_core.GetLoginInfoRequest) (*emi_core.GetLoginInfoResponse, error) {
	var resp emi_core.GetLoginInfoResponse
	if err := h.Post(ctx, string(emi_core.GetLoginInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取协议端信息
func (h *HttpClient) GetImplInfo(ctx context.Context, request emi_core.GetImplInfoRequest) (*emi_core.GetImplInfoResponse, error) {
	var resp emi_core.GetImplInfoResponse
	if err := h.Post(ctx, string(emi_core.GetImplInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取用户个人信息
func (h *HttpClient) GetUserProfile(ctx context.Context, request emi_core.GetUserProfileRequest) (*emi_core.GetUserProfileResponse, error) {
	var resp emi_core.GetUserProfileResponse
	if err := h.Post(ctx, string(emi_core.GetUserProfile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取好友列表
func (h *HttpClient) GetFriendList(ctx context.Context, request emi_core.GetFriendListRequest) (*emi_core.GetFriendListResponse, error) {
	var resp emi_core.GetFriendListResponse
	if err := h.Post(ctx, string(emi_core.GetFriendList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取好友信息
func (h *HttpClient) GetFriendInfo(ctx context.Context, request emi_core.GetFriendInfoRequest) (*emi_core.GetFriendInfoResponse, error) {
	var resp emi_core.GetFriendInfoResponse
	if err := h.Post(ctx, string(emi_core.GetFriendInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群列表
func (h *HttpClient) GetGroupList(ctx context.Context, request emi_core.GetGroupListRequest) (*emi_core.GetGroupListResponse, error) {
	var resp emi_core.GetGroupListResponse
	if err := h.Post(ctx, string(emi_core.GetGroupList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群信息
func (h *HttpClient) GetGroupInfo(ctx context.Context, request emi_core.GetGroupInfoRequest) (*emi_core.GetGroupInfoResponse, error) {
	var resp emi_core.GetGroupInfoResponse
	if err := h.Post(ctx, string(emi_core.GetGroupInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群成员列表
func (h *HttpClient) GetGroupMemberList(ctx context.Context, request emi_core.GetGroupMemberListRequest) (*emi_core.GetGroupMemberListResponse, error) {
	var resp emi_core.GetGroupMemberListResponse
	if err := h.Post(ctx, string(emi_core.GetGroupMemberList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群成员信息
func (h *HttpClient) GetGroupMemberInfo(ctx context.Context, request emi_core.GetGroupMemberInfoRequest) (*emi_core.GetGroupMemberInfoResponse, error) {
	var resp emi_core.GetGroupMemberInfoResponse
	if err := h.Post(ctx, string(emi_core.GetGroupMemberInfo), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置 QQ 账号头像
func (h *HttpClient) SetAvatar(ctx context.Context, request emi_core.SetAvatarRequest) (*emi_core.SetAvatarResponse, error) {
	var resp emi_core.SetAvatarResponse
	if err := h.Post(ctx, string(emi_core.SetAvatar), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置 QQ 账号昵称
func (h *HttpClient) SetNickname(ctx context.Context, request emi_core.SetNicknameRequest) (*emi_core.SetNicknameResponse, error) {
	var resp emi_core.SetNicknameResponse
	if err := h.Post(ctx, string(emi_core.SetNickname), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置 QQ 账号个性签名
func (h *HttpClient) SetBio(ctx context.Context, request emi_core.SetBioRequest) (*emi_core.SetBioResponse, error) {
	var resp emi_core.SetBioResponse
	if err := h.Post(ctx, string(emi_core.SetBio), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取自定义表情 URL 列表
func (h *HttpClient) GetCustomFaceURLList(ctx context.Context, request emi_core.GetCustomFaceURLListRequest) (*emi_core.GetCustomFaceURLListResponse, error) {
	var resp emi_core.GetCustomFaceURLListResponse
	if err := h.Post(ctx, string(emi_core.GetCustomFaceURLList), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取 Cookies
func (h *HttpClient) GetCookies(ctx context.Context, request emi_core.GetCookiesRequest) (*emi_core.GetCookiesResponse, error) {
	var resp emi_core.GetCookiesResponse
	if err := h.Post(ctx, string(emi_core.GetCookies), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取 CSRF Token
func (h *HttpClient) GetCSRFToken(ctx context.Context, request emi_core.GetCSRFTokenRequest) (*emi_core.GetCSRFTokenResponse, error) {
	var resp emi_core.GetCSRFTokenResponse
	if err := h.Post(ctx, string(emi_core.GetCSRFToken), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MessageAPI

// 发送私聊消息
func (h *HttpClient) SendPrivateMessage(ctx context.Context, request emi_core.SendPrivateMessageRequest) (*emi_core.SendPrivateMessageResponse, error) {
	var resp emi_core.SendPrivateMessageResponse
	if err := h.Post(ctx, string(emi_core.SendPrivateMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群聊消息
func (h *HttpClient) SendGroupMessage(ctx context.Context, request emi_core.SendGroupMessageRequest) (*emi_core.SendGroupMessageResponse, error) {
	var resp emi_core.SendGroupMessageResponse
	if err := h.Post(ctx, string(emi_core.SendGroupMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 撤回私聊消息
func (h *HttpClient) RecallPrivateMessage(ctx context.Context, request emi_core.RecallPrivateMessageRequest) (*emi_core.RecallPrivateMessageResponse, error) {
	var resp emi_core.RecallPrivateMessageResponse
	if err := h.Post(ctx, string(emi_core.RecallPrivateMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 撤回群聊消息
func (h *HttpClient) RecallGroupMessage(ctx context.Context, request emi_core.RecallGroupMessageRequest) (*emi_core.RecallGroupMessageResponse, error) {
	var resp emi_core.RecallGroupMessageResponse
	if err := h.Post(ctx, string(emi_core.RecallGroupMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取消息
func (h *HttpClient) GetMessage(ctx context.Context, request emi_core.GetMessageRequest) (*emi_core.GetMessageResponse, error) {
	var resp emi_core.GetMessageResponse
	if err := h.Post(ctx, string(emi_core.GetMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取历史消息列表
func (h *HttpClient) GetHistoryMessages(ctx context.Context, request emi_core.GetHistoryMessagesRequest) (*emi_core.GetHistoryMessagesResponse, error) {
	var resp emi_core.GetHistoryMessagesResponse
	if err := h.Post(ctx, string(emi_core.GetHistoryMessages), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取临时资源链接
func (h *HttpClient) GetResourceTempURL(ctx context.Context, request emi_core.GetResourceTempURLRequest) (*emi_core.GetResourceTempURLResponse, error) {
	var resp emi_core.GetResourceTempURLResponse
	if err := h.Post(ctx, string(emi_core.GetResourceTempURL), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取合并转发消息内容
func (h *HttpClient) GetForwardedMessages(ctx context.Context, request emi_core.GetForwardedMessagesRequest) (*emi_core.GetForwardedMessagesResponse, error) {
	var resp emi_core.GetForwardedMessagesResponse
	if err := h.Post(ctx, string(emi_core.GetForwardedMessages), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 标记消息为已读
func (h *HttpClient) MarkMessageAsRead(ctx context.Context, request emi_core.MarkMessageAsReadRequest) (*emi_core.MarkMessageAsReadResponse, error) {
	var resp emi_core.MarkMessageAsReadResponse
	if err := h.Post(ctx, string(emi_core.MarkMessageAsRead), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FriendAPI

// 发送好友戳一戳
func (h *HttpClient) SendFriendNudge(ctx context.Context, request emi_core.SendFriendNudgeRequest) (*emi_core.SendFriendNudgeResponse, error) {
	var resp emi_core.SendFriendNudgeResponse
	if err := h.Post(ctx, string(emi_core.SendFriendNudge), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送名片点赞
func (h *HttpClient) SendProfileLike(ctx context.Context, request emi_core.SendProfileLikeRequest) (*emi_core.SendProfileLikeResponse, error) {
	var resp emi_core.SendProfileLikeResponse
	if err := h.Post(ctx, string(emi_core.SendProfileLike), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除好友
func (h *HttpClient) DeleteFriend(ctx context.Context, request emi_core.DeleteFriendRequest) (*emi_core.DeleteFriendResponse, error) {
	var resp emi_core.DeleteFriendResponse
	if err := h.Post(ctx, string(emi_core.DeleteFriend), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取好友请求列表
func (h *HttpClient) GetFriendRequests(ctx context.Context, request emi_core.GetFriendRequestsRequest) (*emi_core.GetFriendRequestsResponse, error) {
	var resp emi_core.GetFriendRequestsResponse
	if err := h.Post(ctx, string(emi_core.GetFriendRequests), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 同意好友请求
func (h *HttpClient) AcceptFriendRequest(ctx context.Context, request emi_core.AcceptFriendRequestRequest) (*emi_core.AcceptFriendRequestResponse, error) {
	var resp emi_core.AcceptFriendRequestResponse
	if err := h.Post(ctx, string(emi_core.AcceptFriendRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 拒绝好友请求
func (h *HttpClient) RejectFriendRequest(ctx context.Context, request emi_core.RejectFriendRequestRequest) (*emi_core.RejectFriendRequestResponse, error) {
	var resp emi_core.RejectFriendRequestResponse
	if err := h.Post(ctx, string(emi_core.RejectFriendRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GroupAPI

// 设置群名称
func (h *HttpClient) SetGroupName(ctx context.Context, request emi_core.SetGroupNameRequest) (*emi_core.SetGroupNameResponse, error) {
	var resp emi_core.SetGroupNameResponse
	if err := h.Post(ctx, string(emi_core.SetGroupName), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群头像
func (h *HttpClient) SetGroupAvatar(ctx context.Context, request emi_core.SetGroupAvatarRequest) (*emi_core.SetGroupAvatarResponse, error) {
	var resp emi_core.SetGroupAvatarResponse
	if err := h.Post(ctx, string(emi_core.SetGroupAvatar), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群名片
func (h *HttpClient) SetGroupMemberCard(ctx context.Context, request emi_core.SetGroupMemberCardRequest) (*emi_core.SetGroupMemberCardResponse, error) {
	var resp emi_core.SetGroupMemberCardResponse
	if err := h.Post(ctx, string(emi_core.SetGroupMemberCard), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群成员专属头衔
func (h *HttpClient) SetGroupMemberSpecialTitle(ctx context.Context, request emi_core.SetGroupMemberSpecialTitleRequest) (*emi_core.SetGroupMemberSpecialTitleResponse, error) {
	var resp emi_core.SetGroupMemberSpecialTitleResponse
	if err := h.Post(ctx, string(emi_core.SetGroupMemberSpecialTitle), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群管理员
func (h *HttpClient) SetGroupMemberAdmin(ctx context.Context, request emi_core.SetGroupMemberAdminRequest) (*emi_core.SetGroupMemberAdminResponse, error) {
	var resp emi_core.SetGroupMemberAdminResponse
	if err := h.Post(ctx, string(emi_core.SetGroupMemberAdmin), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群成员禁言
func (h *HttpClient) SetGroupMemberMute(ctx context.Context, request emi_core.SetGroupMemberMuteRequest) (*emi_core.SetGroupMemberMuteResponse, error) {
	var resp emi_core.SetGroupMemberMuteResponse
	if err := h.Post(ctx, string(emi_core.SetGroupMemberMute), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群全员禁言
func (h *HttpClient) SetGroupMemberWholeMute(ctx context.Context, request emi_core.SetGroupMemberWholeMuteRequest) (*emi_core.SetGroupMemberWholeMuteResponse, error) {
	var resp emi_core.SetGroupMemberWholeMuteResponse
	if err := h.Post(ctx, string(emi_core.SetGroupMemberWholeMute), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 踢出群成员
func (h *HttpClient) KickGroupMember(ctx context.Context, request emi_core.KickGroupMemberRequest) (*emi_core.KickGroupMemberResponse, error) {
	var resp emi_core.KickGroupMemberResponse
	if err := h.Post(ctx, string(emi_core.KickGroupMember), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群公告列表
func (h *HttpClient) GetGroupAnnouncements(ctx context.Context, request emi_core.GetGroupAnnouncementsRequest) (*emi_core.GetGroupAnnouncementsResponse, error) {
	var resp emi_core.GetGroupAnnouncementsResponse
	if err := h.Post(ctx, string(emi_core.GetGroupAnnouncements), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群公告
func (h *HttpClient) SendGroupAnnouncement(ctx context.Context, request emi_core.SendGroupAnnouncementRequest) (*emi_core.SendGroupAnnouncementResponse, error) {
	var resp emi_core.SendGroupAnnouncementResponse
	if err := h.Post(ctx, string(emi_core.SendGroupAnnouncement), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除群公告
func (h *HttpClient) DeleteGroupAnnouncement(ctx context.Context, request emi_core.DeleteGroupAnnouncementRequest) (*emi_core.DeleteGroupAnnouncementResponse, error) {
	var resp emi_core.DeleteGroupAnnouncementResponse
	if err := h.Post(ctx, string(emi_core.DeleteGroupAnnouncement), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群精华消息列表
func (h *HttpClient) GetGroupEssenceMessages(ctx context.Context, request emi_core.GetGroupEssenceMessagesRequest) (*emi_core.GetGroupEssenceMessagesResponse, error) {
	var resp emi_core.GetGroupEssenceMessagesResponse
	if err := h.Post(ctx, string(emi_core.GetGroupEssenceMessages), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 设置群精华消息
func (h *HttpClient) SetGroupEssenceMessage(ctx context.Context, request emi_core.SetGroupEssenceMessageRequest) (*emi_core.SetGroupEssenceMessageResponse, error) {
	var resp emi_core.SetGroupEssenceMessageResponse
	if err := h.Post(ctx, string(emi_core.SetGroupEssenceMessage), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 退出群
func (h *HttpClient) QuitGroup(ctx context.Context, request emi_core.QuitGroupRequest) (*emi_core.QuitGroupResponse, error) {
	var resp emi_core.QuitGroupResponse
	if err := h.Post(ctx, string(emi_core.QuitGroup), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群消息表情回应
func (h *HttpClient) SendGroupMessageReaction(ctx context.Context, request emi_core.SendGroupMessageReactionRequest) (*emi_core.SendGroupMessageReactionResponse, error) {
	var resp emi_core.SendGroupMessageReactionResponse
	if err := h.Post(ctx, string(emi_core.SendGroupMessageReaction), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 发送群戳一戳
func (h *HttpClient) SendGroupNudge(ctx context.Context, request emi_core.SendGroupNudgeRequest) (*emi_core.SendGroupNudgeResponse, error) {
	var resp emi_core.SendGroupNudgeResponse
	if err := h.Post(ctx, string(emi_core.SendGroupNudge), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群通知列表
func (h *HttpClient) GetGroupNotifications(ctx context.Context, request emi_core.GetGroupNotificationsRequest) (*emi_core.GetGroupNotificationsResponse, error) {
	var resp emi_core.GetGroupNotificationsResponse
	if err := h.Post(ctx, string(emi_core.GetGroupNotifications), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 同意入群/邀请他人入群请求
func (h *HttpClient) AcceptGroupRequest(ctx context.Context, request emi_core.AcceptGroupRequestRequest) (*emi_core.AcceptGroupRequestResponse, error) {
	var resp emi_core.AcceptGroupRequestResponse
	if err := h.Post(ctx, string(emi_core.AcceptGroupRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 拒绝入群/邀请他人入群请求
func (h *HttpClient) RejectGroupRequest(ctx context.Context, request emi_core.RejectGroupRequestRequest) (*emi_core.RejectGroupRequestResponse, error) {
	var resp emi_core.RejectGroupRequestResponse
	if err := h.Post(ctx, string(emi_core.RejectGroupRequest), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 同意他人邀请自身入群
func (h *HttpClient) AcceptGroupInvitation(ctx context.Context, request emi_core.AcceptGroupInvitationRequest) (*emi_core.AcceptGroupInvitationResponse, error) {
	var resp emi_core.AcceptGroupInvitationResponse
	if err := h.Post(ctx, string(emi_core.AcceptGroupInvitation), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 拒绝他人邀请自身入群
func (h *HttpClient) RejectGroupInvitation(ctx context.Context, request emi_core.RejectGroupInvitationRequest) (*emi_core.RejectGroupInvitationResponse, error) {
	var resp emi_core.RejectGroupInvitationResponse
	if err := h.Post(ctx, string(emi_core.RejectGroupInvitation), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FileAPI

// 上传私聊文件
func (h *HttpClient) UploadPrivateFile(ctx context.Context, request emi_core.UploadPrivateFileRequest) (*emi_core.UploadPrivateFileResponse, error) {
	var resp emi_core.UploadPrivateFileResponse
	if err := h.Post(ctx, string(emi_core.UploadPrivateFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 上传群文件
func (h *HttpClient) UploadGroupFile(ctx context.Context, request emi_core.UploadGroupFileRequest) (*emi_core.UploadGroupFileResponse, error) {
	var resp emi_core.UploadGroupFileResponse
	if err := h.Post(ctx, string(emi_core.UploadGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取私聊文件下载链接
func (h *HttpClient) GetPrivateFileDownloadURL(ctx context.Context, request emi_core.GetPrivateFileDownloadURLRequest) (*emi_core.GetPrivateFileDownloadURLResponse, error) {
	var resp emi_core.GetPrivateFileDownloadURLResponse
	if err := h.Post(ctx, string(emi_core.GetPrivateFileDownloadURL), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群文件下载链接
func (h *HttpClient) GetGroupFileDownloadURL(ctx context.Context, request emi_core.GetGroupFileDownloadURLRequest) (*emi_core.GetGroupFileDownloadURLResponse, error) {
	var resp emi_core.GetGroupFileDownloadURLResponse
	if err := h.Post(ctx, string(emi_core.GetGroupFileDownloadURL), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 获取群文件列表
func (h *HttpClient) GetGroupFiles(ctx context.Context, request emi_core.GetGroupFilesRequest) (*emi_core.GetGroupFilesResponse, error) {
	var resp emi_core.GetGroupFilesResponse
	if err := h.Post(ctx, string(emi_core.GetGroupFiles), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 移动群文件
func (h *HttpClient) MoveGroupFile(ctx context.Context, request emi_core.MoveGroupFileRequest) (*emi_core.MoveGroupFileResponse, error) {
	var resp emi_core.MoveGroupFileResponse
	if err := h.Post(ctx, string(emi_core.MoveGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 重命名群文件
func (h *HttpClient) RenameGroupFile(ctx context.Context, request emi_core.RenameGroupFileRequest) (*emi_core.RenameGroupFileResponse, error) {
	var resp emi_core.RenameGroupFileResponse
	if err := h.Post(ctx, string(emi_core.RenameGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除群文件
func (h *HttpClient) DeleteGroupFile(ctx context.Context, request emi_core.DeleteGroupFileRequest) (*emi_core.DeleteGroupFileResponse, error) {
	var resp emi_core.DeleteGroupFileResponse
	if err := h.Post(ctx, string(emi_core.DeleteGroupFile), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 创建群文件夹
func (h *HttpClient) CreateGroupFolder(ctx context.Context, request emi_core.CreateGroupFolderRequest) (*emi_core.CreateGroupFolderResponse, error) {
	var resp emi_core.CreateGroupFolderResponse
	if err := h.Post(ctx, string(emi_core.CreateGroupFolder), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 重命名群文件夹
func (h *HttpClient) RenameGroupFolder(ctx context.Context, request emi_core.RenameGroupFolderRequest) (*emi_core.RenameGroupFolderResponse, error) {
	var resp emi_core.RenameGroupFolderResponse
	if err := h.Post(ctx, string(emi_core.RenameGroupFolder), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// 删除群文件夹
func (h *HttpClient) DeleteGroupFolder(ctx context.Context, request emi_core.DeleteGroupFolderRequest) (*emi_core.DeleteGroupFolderResponse, error) {
	var resp emi_core.DeleteGroupFolderResponse
	if err := h.Post(ctx, string(emi_core.DeleteGroupFolder), request, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
