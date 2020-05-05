package youtubeapi

type ChannelLiveStreamDetails struct {
	Items []ChannelItems `json:"items"`
}

type ChannelItems struct {
	ID ChannelId `json:"id"`
}

type ChannelId struct {
	VideoId string `json:"videoId"`
}

type LiveStreamDetails struct {
	Items []LiveStreamItems `json:"items"`
}

type LiveStreamItems struct {
	Details LiveStreamingDetails `json:"liveStreamingDetails"`
}

type LiveStreamingDetails struct {
	LiveChatId string `json:"activeLiveChatId"`
}

type CommentToPost struct {
	Kind    string               `json:"kind"`
	Snippet CommentToPostSnippet `json:"snippet"`
}

type CommentToPostSnippet struct {
	Type       string               `json:"type"`
	ChatId     string               `json:"liveChatId"`
	AuthorId   string               `json:"authorChannelId"`
	DisplayMsg string               `json:"displayMessage"`
	Details    CommentToPostDetails `json:"textMessageDetails"`
}

type CommentToPostDetails struct {
	Text string `json:"messageText"`
}

type MessageResponse struct {
	Next     string        `json:"nextPageToken"`
	Messages []MessageItem `json:"items"`
	Info     PageInfo      `json:"pageInfo"`
}

type PageInfo struct {
	Total   int `json:"totalResults"`
	PerPage int `json:"resultsPerPage"`
}

type MessageItem struct {
	Snippet MessageSnippet `json:"snippet"`
}

type MessageSnippet struct {
	Type           string         `json:"type"`
	Author         string         `json:"authorChannelId"`
	DisplayContent bool           `json:"hasDisplayContent"`
	DisplayMessage string         `json:"displayMessage"`
	Details        MessageDetails `json:"textMessageDetails"`
}

type MessageDetails struct {
	Text string `json:"messageText"`
}

type UserFromChannelResponse struct {
	Items []UserItems `json:"items"`
}

type UserItems struct {
	Snippet UserSnippet `json:"snippet"`
}

type UserSnippet struct {
	Local Localized `json:"localized"`
}

type Localized struct {
	Title string `json:"title"`
}

type TokenResponse struct {
	Token      string `json:"access_token"`
	Expiration int    `json:"expires_in"`
}
