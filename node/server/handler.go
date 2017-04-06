package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
	"xsbPro/chat/lua"
	"xsbPro/chat/node/server/communication"
	"xsbPro/chat/node/server/user/detail"
	"xsbPro/log"
	"xsbPro/xsbAdmin/libs/tools"

	"xsbPro/chat/node/server/hub"
	"xsbPro/chat/node/server/user"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// pingPeriod = 5 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 4
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NodeStatusReport returns a node status report
func NodeStatusReport() *SummaryReport {
	// sr := MakeHubsStatusReport(hubManager.GetHubs())
	// return sr
	return nil
}

// // checkSameOrigin returns true if the origin is not set or is equal to the request host.
// func checkOriginAllowed(r *http.Request) bool {
// 	return true
// }

// NewUserRequest handles new connect request
// groupID is hub id
func NewUserRequest(groupID, userID string, c *gin.Context, scriptExecutor lua.ScriptExecutor) error {
	hubUser, err := loadUser(userID, groupID, nodeID, scriptExecutor)
	if err != nil {
		return err
	}
	if hubUser == nil {
		log.InfoF("no user %s in group %s", userID, groupID)
		return fmt.Errorf("noServiceForThisGroup")
	}

	log.TraceF("user (%s %s) in group (%s) comes in", hubUser.GetID(), hubUser.GetName(), groupID)

	ws, err := upgradeToWebsocket(c.Writer, c.Request, nil)
	if err != nil {
		log.SysF(err.Error())
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	hubUser.SetConn(ws, cancel)
	log.TraceF("user %s refresh connection", hubUser.GetID())
	<-ctx.Done()
	log.TraceF("user %s leave", hubUser.GetID())
	return nil
}

func loadUser(userID, groupID, nodeID string, scriptExecutor lua.ScriptExecutor) (u *user.User, err error) {
	hub, err := loadHub(groupID, nodeID, scriptExecutor)
	if err != nil {
		return nil, err
	}
	// if hub is still nil, it means this node should not load this group
	if hub == nil {
		return nil, nil
	}

	u = hub.FindUser(userID)
	if u != nil {
		return
	}

	// if no user found, then
	if strings.HasPrefix(userID, "iamafakeuser") { //以虚假用户身份登录
		fakeUser := detail.NewFakeUser(userID)
		u = user.NewUser(fakeUser, hub)
		hub.AddUser(u)
		return
	}
	return
}

func loadHub(group, node string, scriptExecutor lua.ScriptExecutor) (h *hub.Hub, err error) {
	h = serverInstance.findHub(group)
	if h == nil {
		h, err = loadGroupFromRedis(group, node, scriptExecutor)
		if err != nil {
			return
		}
	}
	return
}

func upgradeToWebsocket(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {

	ws, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	return ws, nil
}

// ImageUpload handlers image message
func ImageUpload(groupID, userID, para string, c *gin.Context) {

	hub := serverInstance.findHub(groupID)
	if hub == nil {
		log.SysF("group %s does NOT exists", groupID)
		c.AbortWithError(http.StatusOK, errors.New("no group ID"))
		return
	}

	ui := hub.FindUser(userID)
	if ui == nil {
		log.SysF("user %s does NOT exists", userID)
		c.AbortWithError(http.StatusBadRequest, errors.New("no user ID"))
		return
	}

	log.TraceF("UploadImage -> uid: %s %s", userID, ui.GetName())

	// para := c.Query("para")

	header1, err := getUploadFileHeader(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
		return
	}
	f, err := header1.Open()
	if err != nil {
		log.InfoF("open uploaded file error: %s", err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
		return
	}

	res, body, err := tools.UploadFile(fmt.Sprintf(uploadStaticImageFileURL, userID, para), header1.Filename, nil, f)
	if err != nil {
		log.InfoF("upload image to server error: %s", err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
		return
	}
	log.TraceF("upload avatar to statics server: %s : %s", res.Status, string(body))

	type ImageUploadResponse struct {
		State string `json:"state"`
		URL   string `json:"url"`
	}
	var r ImageUploadResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.InfoF("Unmarshal json error: %s", err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
		return
	}

	if len(r.URL) <= 0 {
		log.InfoF("upload image failed")
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
		return
	}

	message, err := communication.NewImageMessage(ui.GetID(), ui.GetName(), r.URL)
	if err != nil {
		log.SysF("user %s add Image message error: %s", userID, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = ui.NewMessage(message)
	// err = hub.NewMessage(message, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.TraceF("user %s add IMAGE message", userID)

	c.JSON(http.StatusOK, nil)
}

// AudioUpload handles audio msg
func AudioUpload(groupID, userID, para string, c *gin.Context) {

	hub := serverInstance.findHub(groupID)
	if hub == nil {
		log.SysF("group %s does NOT exists", groupID)
		c.AbortWithError(http.StatusOK, errors.New("no group ID"))
		return
	}

	ui := hub.FindUser(userID)
	if ui == nil {
		log.SysF("user %s does NOT exists", userID)
		c.AbortWithError(http.StatusBadRequest, errors.New("no user ID"))
		return
	}

	// log.TraceF("UploadAudio -> userID %s  para: %s ", userID, para)

	// log.TraceF("user %s UploadAudio", ui.Name)
	header1, err := getUploadFileHeader(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
		return
	}

	f, err := header1.Open()
	if err != nil {
		log.InfoF("open uploaded file error: %s", err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
		return
	}

	res, body, err := tools.UploadFile(fmt.Sprintf(uploadStaticAudioFileURL, userID, para), header1.Filename, nil, f)
	if err != nil {
		log.InfoF("upload audio to server error: %s", err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
		return
	}
	log.TraceF("upload audio to statics server: %s : %s", res.Status, string(body))
	// fileName := para + "_" + userID + "_" + strconv.FormatInt(time.Now().UnixNano(), 16) + filepath.Ext(header1.Filename)

	// log.TraceF("create audio fileName: %s", fileName)
	// tofile := path.Join(chatFilesPathAudio, fileName)

	// err = receiveUploadFile(tofile, header1)
	// if err != nil {
	// 	log.SysF("接收语音文件失败:%s", err)
	// 	c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
	// 	return
	// }

	type Response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil || response.Data == nil {
		log.InfoF("Unmarshal json error: %s", err)
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
		return
	}

	if len(response.Data.(string)) <= 0 {
		log.InfoF("upload audio failed")
		c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
		return
	}

	message, err := communication.NewAudioMessage(ui.GetID(), ui.GetName(), response.Data.(string))
	if err != nil {
		log.SysF("发送语音消息失败:%s", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = ui.NewMessage(message)
	// err = hub.NewMessage(message, nil)
	if err != nil {
		log.SysF("发送语音消息失败:%s", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.TraceF("user %s add AUDIO message", userID)

	c.JSON(http.StatusOK, nil)
}

func receiveUploadFile(newFileFullPath string, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}

	defer file.Close()

	f, err := os.OpenFile(newFileFullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	return nil
}

func getUploadFileHeader(req *http.Request) (*multipart.FileHeader, error) {

	err := req.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		log.Sys(err.Error())
		return nil, err
	}

	if req.MultipartForm == nil {
		return nil, errors.New("no file uploaded")
	}

	mf := req.MultipartForm
	if len(mf.File) <= 0 {
		return nil, errors.New("no file uploaded")
	}

	var header1 *multipart.FileHeader

	for _, headers := range mf.File {
		header1 = headers[0]
		break
	}
	return header1, nil
}
