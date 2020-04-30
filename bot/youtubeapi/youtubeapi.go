package youtubeapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var ErrorNilChannelID error = errors.New("Empty Channel ID.")
var ErrorNoApiKey error = errors.New("You need to provide an API key.")
var ErrorApiCall error = errors.New("Error calling the Youtube API")
var ErrorDecoding error = errors.New("There was an error decoding a response.")
var ErrorNilLivestreamID error = errors.New("Empty Livestream ID.")
var ErrorNoActiveLivestreams error = errors.New("Channel has no active livestreams.")
var ErrorNotFound error = errors.New("Resource not found.")

var apiKey string
var token string

const (
	urlLivestreamFromChannel = "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=#UID&eventType=live&type=video&key="
	urlLiveChatId            = "https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails&id=#UID&key="
	urlPostComment           = "https://www.googleapis.com/youtube/v3/liveChat/messages?part=snippet&key="
	urlGetMessages           = "https://www.googleapis.com/youtube/v3/liveChat/messages?liveChatId=#UID&part=snippet&key="
	urlGetUser               = "https://www.googleapis.com/youtube/v3/channels?part=snippet&id=#UID&key="
	urlOauth                 = "https://oauth2.googleapis.com/token?"
	pageToken                = "&pageToken="
	client_id                = "client_id="
	client_secret            = "client_secret="
	refresh_token            = "refresh_token="
	grant_type               = "grant_type=refresh_token"
)

func SetApiKey(s string) {
	apiKey = s
}

func SetToken(s string) {
	token = s
}

func GetLivestreamIdFromChannelId(u string, l *log.Logger) ([]string, error) {
	if u == "" {
		l.Println(ErrorNilChannelID.Error())
		return nil, ErrorNilChannelID
	}
	if apiKey == "" {
		l.Println(ErrorNoApiKey.Error())
		return nil, ErrorNoApiKey
	}
	urlGet := urlLivestreamFromChannel + url.QueryEscape(apiKey)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(u), 1)
	r, err := doGet(urlGet, l)
	if err != nil {
		return nil, err
	}
	var details ChannelLiveStreamDetails
	errD := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		l.Println(errD.Error())
		return nil, ErrorDecoding
	}
	response := make([]string, len(details.Items))
	for i := range details.Items {
		response[i] = details.Items[i].ID.VideoId
	}
	return response, nil

}

func GetLiveChatIdFromLiveStreamId(s string, l *log.Logger) (string, error) {
	if s == "" {
		l.Println(ErrorNilChannelID.Error())
		return "", ErrorNilChannelID
	}
	if apiKey == "" {
		l.Println(ErrorNoApiKey.Error())
		return "", ErrorNoApiKey
	}
	urlGet := urlLiveChatId + url.QueryEscape(apiKey)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(s), 1)
	r, err := doGet(urlGet, l)
	if err != nil {
		return "", err
	}
	var details *LiveStreamDetails
	errD := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		l.Println(errD.Error())
		return "", ErrorDecoding
	}
	return details.Items[0].Details.LiveChatId, nil

}

func GetFristLiveChatIdFromChannelId(c string, l *log.Logger) (string, error) {
	ids, err := GetLivestreamIdFromChannelId(c, l)
	if err != nil {
		return "", err
	}
	if len(ids) < 1 {
		return "", ErrorNoActiveLivestreams
	}
	liveChatId, err2 := GetLiveChatIdFromLiveStreamId(ids[0], l)
	if err2 != nil {
		return "", err
	}
	return liveChatId, nil
}

func PostComment(message string, chatId string, author string, l *log.Logger) error {
	if message == "" {
		return errors.New("Yo cant post an empty comment.")
	}
	urlPost := urlPostComment + url.QueryEscape(apiKey)
	payload := NewCommentToPost(chatId, author, message)
	bytesM, errM := json.Marshal(payload)
	if errM != nil {
		return errors.New("Cant encode message to post.")
	}
	err := doPostWithOauth2(urlPost, bytesM, l)
	if err != nil {
		return err
	}
	return nil

}

func ReadMessages(c string, n string, l *log.Logger) (MessageResponse, error) {
	if c == "" {
		return MessageResponse{}, ErrorNilChannelID
	}
	if apiKey == "" {
		l.Println(ErrorNoApiKey.Error())
		return MessageResponse{}, ErrorNoApiKey
	}
	urlGet := urlGetMessages + url.QueryEscape(apiKey)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(c), 1)
	if n != "" {
		urlGet = urlGet + pageToken + url.QueryEscape(n)
	}
	r, err := doGet(urlGet, l)
	if err != nil {
		return MessageResponse{}, err
	}
	var messages MessageResponse
	errD := json.NewDecoder(r.Body).Decode(&messages)
	if err != nil {
		l.Println(errD.Error())
		return MessageResponse{}, ErrorDecoding
	}
	return messages, nil
}

func GetUserFromChannelId(c string, l *log.Logger) (string, error) {
	if c == "" {
		return "", ErrorNilChannelID
	}
	if apiKey == "" {
		l.Println(ErrorNoApiKey.Error())
		return "", ErrorNoApiKey
	}
	urlGet := urlGetUser + url.QueryEscape(apiKey)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(c), 1)
	l.Println(urlGet)
	r, err := doGet(urlGet, l)
	if err != nil {
		return "", err
	}
	var user UserFromChannelResponse
	errD := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		l.Println(errD.Error())
		return "", ErrorDecoding
	}
	if len(user.Items) < 1 {
		return "", ErrorNotFound
	}
	l.Println(user.Items[0].Snippet.Local.Title)
	return user.Items[0].Snippet.Local.Title, nil
}

func GetNewAuthToken(cId string, cSec string, ref string, l *log.Logger) (string, error) {
	if cId == "" || cSec == "" || ref == "" {
		return "", errors.New("Error missing data needed for new token.")
	}
	urlPost := urlOauth + client_id + url.QueryEscape(cId) + "&" + client_secret + url.QueryEscape(cSec) + "&" + refresh_token + url.QueryEscape(ref) + "&" + grant_type
	res, err := doPost(urlPost, make([]byte, 0), nil, l)
	if err != nil {
		return "", err
	}
	var token TokenResponse
	json.NewDecoder(res.Body).Decode(&token)
	return token.Token, nil
}

func doGet(u string, l *log.Logger) (*http.Response, error) {
	r, err := http.Get(u)
	if err != nil || r.StatusCode != 200 {
		l.Println(err.Error())
		return nil, ErrorApiCall
	} else {
		return r, nil
	}
}

func doPostWithOauth2(u string, p []byte, l *log.Logger) error {
	head := make(map[string]string)
	head["Authorization"] = "Bearer " + token
	_, err := doPost(u, p, head, l)
	if err != nil {
		return err
	}
	return nil
}

func doPost(u string, p []byte, head map[string]string, l *log.Logger) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", u, bytes.NewBuffer(p))
	for k, v := range head {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		l.Println(err.Error())
		return nil, ErrorApiCall
	} else if res.StatusCode != 200 {
		bs, _ := ioutil.ReadAll(res.Body)
		l.Println(string(bs))
		return nil, ErrorApiCall
	} else {
		return res, nil
	}
}

func NewCommentToPost(l string, a string, m string) CommentToPost {
	d := CommentToPostDetails{m}
	s := CommentToPostSnippet{"textMessageEvent", l, a, "textMessage", d}
	return CommentToPost{"youtube#liveChatMessage", s}
}
