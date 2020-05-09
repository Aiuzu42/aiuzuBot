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
var ErrorNilCommentID error = errors.New("Comment Id not provided")
var ErrorEncoding error = errors.New("Error encoding data.")
var ErrUnauthorized error = errors.New("Unauthorized, invalid credentials")

const (
	urlLivestreamFromChannel = "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=#UID&eventType=live&type=video&key="
	urlLiveChatId            = "https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails&id=#UID&key="
	urlPostComment           = "https://www.googleapis.com/youtube/v3/liveChat/messages?part=snippet&key="
	urlGetMessages           = "https://www.googleapis.com/youtube/v3/liveChat/messages?liveChatId=#UID&part=snippet&key="
	urlGetUser               = "https://www.googleapis.com/youtube/v3/channels?part=snippet&id=#UID&key="
	urlDeleteComment         = "https://www.googleapis.com/youtube/v3/liveChat/messages?id=#UID&key="
	urlBanUser               = "https://www.googleapis.com/youtube/v3/liveChat/bans?part=snippet&key="
	urlOauth                 = "https://oauth2.googleapis.com/token?"
	pageToken                = "&pageToken="
	client_id                = "client_id="
	client_secret            = "client_secret="
	refresh_token            = "refresh_token="
	grant_type               = "grant_type=refresh_token"
	temporary_ban            = "temporary"
	permanent_ban            = "permanent"
)

func GetLivestreamIdFromChannelId(u string, key string, l *log.Logger) ([]string, error) {
	if u == "" {
		l.Println(ErrorNilChannelID.Error())
		return nil, ErrorNilChannelID
	}
	if key == "" {
		l.Println(ErrorNoApiKey.Error())
		return nil, ErrorNoApiKey
	}
	urlGet := urlLivestreamFromChannel + url.QueryEscape(key)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(u), 1)
	r, err := doGet(urlGet, l)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var details ChannelLiveStreamDetails
	errD := json.NewDecoder(r.Body).Decode(&details)
	if errD != nil {
		l.Println(errD.Error())
		return nil, ErrorDecoding
	}
	response := make([]string, len(details.Items))
	for i := range details.Items {
		response[i] = details.Items[i].ID.VideoId
	}
	return response, nil

}

func GetLiveChatIdFromLiveStreamId(s string, key string, l *log.Logger) (string, error) {
	if s == "" {
		l.Println(ErrorNilChannelID.Error())
		return "", ErrorNilChannelID
	}
	if key == "" {
		l.Println(ErrorNoApiKey.Error())
		return "", ErrorNoApiKey
	}
	urlGet := urlLiveChatId + url.QueryEscape(key)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(s), 1)
	r, err := doGet(urlGet, l)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	var details *LiveStreamDetails
	errD := json.NewDecoder(r.Body).Decode(&details)
	if errD != nil {
		l.Println(errD.Error())
		return "", ErrorDecoding
	}
	return details.Items[0].Details.LiveChatId, nil

}

func GetFristLiveChatIdFromChannelId(c string, key string, l *log.Logger) (string, error) {
	ids, err := GetLivestreamIdFromChannelId(c, key, l)
	if err != nil {
		return "", err
	}
	if len(ids) < 1 {
		l.Println(ErrorNoActiveLivestreams.Error())
		return "", ErrorNoActiveLivestreams
	}
	liveChatId, err2 := GetLiveChatIdFromLiveStreamId(ids[0], key, l)
	if err2 != nil {
		return "", err
	}
	return liveChatId, nil
}

func PostComment(message string, chatId string, author string, key string, token string, l *log.Logger) error {
	if message == "" {
		return errors.New("Yo cant post an empty comment.")
	}
	urlPost := urlPostComment + url.QueryEscape(key)
	payload := NewCommentToPost(chatId, author, message)
	bytesM, errM := json.Marshal(payload)
	if errM != nil {
		l.Println(ErrorEncoding.Error())
		return ErrorEncoding
	}
	_, err := doPostWithOauth2(urlPost, bytesM, token, l)
	if err != nil {
		return err
	}
	return nil

}

func ReadMessages(c string, n string, key string, l *log.Logger) (MessageResponse, error) {
	if c == "" {
		l.Println(ErrorNilChannelID.Error())
		return MessageResponse{}, ErrorNilChannelID
	}
	if key == "" {
		l.Println(ErrorNoApiKey.Error())
		return MessageResponse{}, ErrorNoApiKey
	}
	urlGet := urlGetMessages + url.QueryEscape(key)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(c), 1)
	if n != "" {
		urlGet = urlGet + pageToken + url.QueryEscape(n)
	}
	r, err := doGet(urlGet, l)
	if err != nil {
		return MessageResponse{}, err
	}
	defer r.Body.Close()
	var messages MessageResponse
	errD := json.NewDecoder(r.Body).Decode(&messages)
	if errD != nil {
		l.Println(errD.Error())
		return MessageResponse{}, ErrorDecoding
	}
	return messages, nil
}

func GetUserFromChannelId(c string, key string, l *log.Logger) (string, error) {
	if c == "" {
		l.Println(ErrorNilChannelID.Error())
		return "", ErrorNilChannelID
	}
	if key == "" {
		l.Println(ErrorNoApiKey.Error())
		return "", ErrorNoApiKey
	}
	urlGet := urlGetUser + url.QueryEscape(key)
	urlGet = strings.Replace(urlGet, "#UID", url.QueryEscape(c), 1)
	r, err := doGet(urlGet, l)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	var user UserFromChannelResponse
	errD := json.NewDecoder(r.Body).Decode(&user)
	if errD != nil {
		l.Println(errD.Error())
		return "", ErrorDecoding
	}
	if len(user.Items) == 0 {
		l.Println("The user was not found.")
		return "", ErrorNotFound
	}
	return user.Items[0].Snippet.Local.Title, nil
}

func DeleteCommment(cId string, key string, token string, l *log.Logger) error {
	if cId == "" {
		l.Println(ErrorNilCommentID.Error())
		return ErrorNilCommentID
	}
	if key == "" {
		l.Println(ErrorNoApiKey.Error())
		return ErrorNoApiKey
	}
	urlDelete := urlDeleteComment + url.QueryEscape(key)
	urlDelete = strings.Replace(urlDelete, "#UID", url.QueryEscape(cId), 1)
	err := doDeleteWithOauth2(urlDelete, token, l)
	if err != nil {
		return err
	}
	return nil
}

func BanUser(chatId string, t string, userId string, d int, key string, token string, l *log.Logger) (string, error) {
	if chatId == "" {
		l.Println(ErrorNilLivestreamID.Error())
		return "", ErrorNilLivestreamID
	}
	if t == "" || (t != temporary_ban && t != permanent_ban) {
		l.Println("Ban type ins incorrect")
		return "", errors.New("Ban type is incorrect")
	}
	if userId == "" {
		l.Println(ErrorNilChannelID.Error())
		return "", ErrorNilChannelID
	}
	urlPost := urlBanUser + url.QueryEscape(key)
	payload := NewBanResource(chatId, t, d, userId)
	bytesM, errM := json.Marshal(payload)
	if errM != nil {
		l.Println(ErrorEncoding.Error())
		return "", ErrorEncoding
	}
	r, err := doPostWithOauth2(urlPost, bytesM, token, l)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	var ban BanResource
	errD := json.NewDecoder(r.Body).Decode(&ban)
	if errD != nil {
		l.Println(ErrorDecoding.Error())
		return "", errD
	}
	return ban.Id, nil
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
	defer res.Body.Close()
	var token TokenResponse
	errD := json.NewDecoder(res.Body).Decode(&token)
	if errD != nil {
		l.Println(errD.Error())
		return "", ErrorDecoding
	}
	return token.Token, nil
}

func doGet(u string, l *log.Logger) (*http.Response, error) {
	r, err := http.Get(u)
	err = handleResponse(r, err, l)
	if err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func doPostWithOauth2(u string, p []byte, token string, l *log.Logger) (*http.Response, error) {
	head := make(map[string]string)
	head["Authorization"] = "Bearer " + token
	r, err := doPost(u, p, head, l)
	if err != nil {
		l.Println(token)
		return nil, err
	}
	return r, nil
}

func doPost(u string, p []byte, head map[string]string, l *log.Logger) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", u, bytes.NewBuffer(p))
	for k, v := range head {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	err = handleResponse(res, err, l)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func doDeleteWithOauth2(url string, token string, l *log.Logger) error {
	head := make(map[string]string)
	head["Authorization"] = "Bearer " + token
	_, err := doDelete(url, head, l)
	if err != nil {
		return err
	}
	return nil
}

func doDelete(url string, head map[string]string, l *log.Logger) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", url, nil)
	for k, v := range head {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	err = handleResponse(res, err, l)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func handleResponse(r *http.Response, e error, l *log.Logger) error {
	if e != nil {
		l.Println(e.Error())
		return e
	} else if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	} else {
		bs, _ := ioutil.ReadAll(r.Body)
		l.Println(string(bs))
		if r.StatusCode == 401 {
			return ErrUnauthorized
		} else {
			return ErrorApiCall
		}
	}
}

func NewCommentToPost(l string, a string, m string) CommentToPost {
	d := CommentToPostDetails{m}
	s := CommentToPostSnippet{"textMessageEvent", l, a, "textMessage", d}
	return CommentToPost{"youtube#liveChatMessage", s}
}

func NewBanResource(chatId string, t string, d int, userId string) BanResource {
	banDetails := BannedUserDetails{userId}
	banSnippet := BanSnippet{chatId, t, d, banDetails}
	return BanResource{Kind: "youtube#liveChatBan", Snippet: banSnippet}
}
