package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"xsbPro/chat/node/resource"
	"xsbPro/chat/node/server"
	"xsbPro/log"
)

// const (
// 	// Time allowed to write a message to the peer.
// 	writeWait = 5 * time.Second

// 	// Time allowed to read the next pong message from the peer.
// 	pongWait = 30 * time.Second

// 	// Send pings to peer with this period. Must be less than pongWait.
// 	pingPeriod = (pongWait * 9) / 10
// 	// pingPeriod = 5 * time.Second

// 	// Maximum message size allowed from peer.
// 	maxMessageSize = 1024 * 4
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin:     checkOriginAllowed,
// }

// // checkSameOrigin returns true if the origin is not set or is equal to the request host.
// func checkOriginAllowed(r *http.Request) bool {
// 	return true
// }

// ServeWs handles websocket requests from the peer.
func ServeWs(c *gin.Context) {
	log.TraceF("new ws request: %s", c.Request.URL)
	// r := c.Request
	// origin := r.Header["Origin"]
	// if len(origin) == 0 {
	// 	log.TraceF("orgin is null")
	// }
	// u, err := url.Parse(origin[0])
	// if err != nil {
	// 	log.TraceF("url parse error: %s", err)
	// }
	// log.TraceF("u.Host: %s  r.Host: %s", u.Host, r.Host)

	userID := c.Query("id")
	groupID := c.Query("group")
	if len(groupID) <= 0 || len(userID) <= 0 {
		log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
		c.AbortWithError(http.StatusOK, errors.New("no user ID or group ID"))
		return
	}
	err := server.NewUserRequest(groupID, userID, c, resource.RedisInstance.DoScript)
	if err != nil {
		c.AbortWithError(http.StatusOK, err)
	}
	// hm := hubManager
	// if hm == nil {
	// 	log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
	// 	c.AbortWithError(http.StatusOK, errors.New("no user ID or group ID"))
	// 	return
	// }
	// hub := hm.Hubs.Get(groupID)
	// if hub == nil {
	// 	// //主动发起承载 group 的请求
	// 	// if hm.Hubs.Length() < conf.GetGroupLoadCapability() {
	// 	// 	select {
	// 	// 	case dispatch_request_group_cache <- groupID:
	// 	// 	default:
	// 	// 	}
	// 	// }
	// 	c.AbortWithError(http.StatusOK, errors.New("noService"))
	// 	return
	// }

	// ui := hub.FindUser(userID)
	// if ui == nil {
	// 	var userInfo models.UserInfo
	// 	if strings.HasPrefix(userID, "iamafakeuser") { //以虚假用户身份登录
	// 		userInfo = models.NewFakeUserInfo(userID)
	// 	} else {
	// 		db_user, err := getUserInGroup(groupID, userID)
	// 		if err != nil {
	// 			log.InfoF("user %s does NOT exists in hub %s", userID, groupID)
	// 			c.AbortWithError(http.StatusOK, errors.New("支部或者用户 ID 错误,或者用户不在该支部中"))
	// 			return
	// 		}
	// 		if db_user != nil {
	// 			userInfo = models.NewRealUserInfo(db_user)
	// 		} else {
	// 			log.InfoF("user %s does NOT exists in hub %s", userID, groupID)
	// 			c.AbortWithError(http.StatusOK, errors.New("支部或者用户 ID 错误,或者用户不在该支部中"))
	// 			return
	// 		}
	// 	}

	// 	//如果数据库空闲,可以实时从数据库中获取数据,否则等待下次重试
	// 	if userInfo != nil {
	// 		ui = models.NewUser(userInfo, hub)
	// 		hub.GroupUsers.Set(ui.User.GetUserID(), ui)
	// 	}
	// }

	// log.TraceF("user (%s %s) in group (%s) comes in", ui.User.GetUserID(), ui.User.GetUserName(), groupID)

	// ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	// if err != nil {
	// 	log.SysF(err.Error())
	// 	// c.AbortWithError(http.StatusOK, err)
	// 	return
	// }
	// ws.SetReadLimit(maxMessageSize)
	// ws.SetReadDeadline(time.Now().Add(pongWait))
	// ws.SetPongHandler(func(string) error {
	// 	ws.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })
	// conn := models.NewConnection(ws, ui, ui.User.GetUserID())

	// ui.SetConn(conn)
	// // err = hub.registerConn(conn)
	// // if err != nil {
	// // 	conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
	// // 	c.AbortWithError(http.StatusOK, err)
	// // 	return
	// // }
	// go conn.WritePump(pingPeriod, writeWait)
	// conn.ReadPump()
}

//UploadAudio use
func UploadAudio(c *gin.Context) {

	userID := c.Query("uid")
	groupID := c.Query("group")
	para := c.Query("para")

	if len(groupID) <= 0 || len(userID) <= 0 {
		log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
		c.AbortWithError(http.StatusOK, errors.New("no user ID or group ID"))
		return
	}

	log.InfoF("get id (%s) in group (%s) upload audio, para %s", userID, groupID, para)

	server.AudioUpload(groupID, userID, para, c)
	// hm := hubManager
	// if hm == nil {
	// 	log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
	// 	c.AbortWithError(http.StatusOK, errors.New("no user ID or group ID"))
	// 	return
	// }
	// hub := hm.Hubs.Get(groupID)
	// if hub == nil {
	// 	log.SysF("group %s does NOT exists", groupID)
	// 	c.AbortWithError(http.StatusOK, errors.New("no group ID"))
	// 	return
	// }

	// ui := hub.FindUser(userID)
	// if ui == nil {
	// 	log.SysF("user %s does NOT exists", userID)
	// 	c.AbortWithError(http.StatusBadRequest, errors.New("no user ID"))
	// 	return
	// }

	// // log.TraceF("UploadAudio -> userID %s  para: %s ", userID, para)

	// // log.TraceF("user %s UploadAudio", ui.Name)
	// header1, err := getUploadFileHeader(c.Request)
	// if err != nil {
	// 	c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
	// 	return
	// }

	// f, err := header1.Open()
	// if err != nil {
	// 	log.InfoF("open uploaded file error: %s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
	// 	return
	// }

	// res, body, err := tools.UploadFile(fmt.Sprintf(upload_static_audio_file_url, userID, para), header1.Filename, nil, f)
	// if err != nil {
	// 	log.InfoF("upload audio to server error: %s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
	// 	return
	// }
	// log.TraceF("upload audio to statics server: %s : %s", res.Status, string(body))
	// // fileName := para + "_" + userID + "_" + strconv.FormatInt(time.Now().UnixNano(), 16) + filepath.Ext(header1.Filename)

	// // log.TraceF("create audio fileName: %s", fileName)
	// // tofile := path.Join(chatFilesPathAudio, fileName)

	// // err = receiveUploadFile(tofile, header1)
	// // if err != nil {
	// // 	log.SysF("接收语音文件失败:%s", err)
	// // 	c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
	// // 	return
	// // }

	// type Response struct {
	// 	Code    int         `json:"code"`
	// 	Message string      `json:"message"`
	// 	Data    interface{} `json:"data"`
	// }

	// var response Response
	// err = json.Unmarshal(body, &response)
	// if err != nil || response.Data == nil {
	// 	log.InfoF("Unmarshal json error: %s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
	// 	return
	// }

	// if len(response.Data.(string)) <= 0 {
	// 	log.InfoF("upload audio failed")
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传失败"))
	// 	return
	// }

	// message, err := models.NewAudioMessage(ui.User, response.Data.(string))
	// if err != nil {
	// 	log.SysF("发送语音消息失败:%s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	// err = hub.NewMessage(message, nil)
	// if err != nil {
	// 	log.SysF("发送语音消息失败:%s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	// log.TraceF("user %s add AUDIO message", userID)

	// c.JSON(http.StatusOK, nil)
}

// UploadImage use
func UploadImage(c *gin.Context) {
	log.TraceF("UploadImage ->")

	userID := c.Query("uid")
	groupID := c.Query("group")
	if len(groupID) <= 0 || len(userID) <= 0 {
		log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
		c.AbortWithError(http.StatusOK, errors.New("no user ID or group ID"))
		return
	}

	log.InfoF("get id (%s) in group (%s) comes in", userID, groupID)

	para := c.Query("para")

	server.ImageUpload(groupID, userID, para, c)
	// hm := hubManager
	// if hm == nil {
	// 	log.InfoF("no user ID: [%s] group ID: [%s]", userID, groupID)
	// 	c.AbortWithError(http.StatusOK, errors.New("no user ID or group ID"))
	// 	return
	// }
	// hub := hm.Hubs.Get(groupID)
	// if hub == nil {
	// 	log.SysF("group %s does NOT exists", groupID)
	// 	c.AbortWithError(http.StatusOK, errors.New("no group ID"))
	// 	return
	// }

	// ui := hub.FindUser(userID)
	// if ui == nil {
	// 	log.SysF("user %s does NOT exists", userID)
	// 	c.AbortWithError(http.StatusBadRequest, errors.New("no user ID"))
	// 	return
	// }

	// log.TraceF("UploadImage -> uid: %s %s", userID, ui.User.GetUserName())

	// header1, err := getUploadFileHeader(c.Request)
	// if err != nil {
	// 	c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
	// 	return
	// }
	// f, err := header1.Open()
	// if err != nil {
	// 	log.InfoF("open uploaded file error: %s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
	// 	return
	// }

	// res, body, err := tools.UploadFile(fmt.Sprintf(upload_static_image_file_url, userID, para), header1.Filename, nil, f)
	// if err != nil {
	// 	log.InfoF("upload image to server error: %s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
	// 	return
	// }
	// log.TraceF("upload avatar to statics server: %s : %s", res.Status, string(body))

	// type ImageUploadResponse struct {
	// 	State string `json:"state"`
	// 	Url   string `json:"url"`
	// }
	// var r ImageUploadResponse
	// err = json.Unmarshal(body, &r)
	// if err != nil {
	// 	log.InfoF("Unmarshal json error: %s", err)
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
	// 	return
	// }

	// if len(r.Url) <= 0 {
	// 	log.InfoF("upload image failed")
	// 	c.AbortWithError(http.StatusInternalServerError, errors.New("上传图片失败"))
	// 	return
	// }
	// // log.TraceF("UploadImage -> fileName: %s", header1.Filename)

	// // fileName := para + "_" + userID + "_" + strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(header1.Filename)
	// // tofile := path.Join(chatFilesPathImage, fileName)

	// // err = receiveUploadFile(tofile, header1)
	// // if err != nil {
	// // 	c.AbortWithError(http.StatusBadRequest, errors.New("上传文件出错"))
	// // 	return
	// // }
	// message, err := models.NewImageMessage(ui.User, r.Url)
	// // message, err := newImageMessage(ui, fileName)
	// if err != nil {
	// 	log.SysF("user %s add Image message error: %s", userID, err)
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	// err = hub.NewMessage(message, nil)
	// if err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	// log.TraceF("user %s add IMAGE message", userID)

	// c.JSON(http.StatusOK, nil)
}

// func receiveUploadFile(newFileFullPath string, fileHeader *multipart.FileHeader) error {
// 	file, err := fileHeader.Open()
// 	if err != nil {
// 		return err
// 	}

// 	defer file.Close()

// 	f, err := os.OpenFile(newFileFullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
// 	if err != nil {
// 		return err
// 	}

// 	defer f.Close()
// 	_, err = io.Copy(f, file)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func getUploadFileHeader(req *http.Request) (*multipart.FileHeader, error) {

// 	err := req.ParseMultipartForm(defaultMaxMemory)
// 	if err != nil {
// 		log.Sys(err.Error())
// 		return nil, err
// 	}

// 	if req.MultipartForm == nil {
// 		return nil, errors.New("no file uploaded")
// 	}

// 	mf := req.MultipartForm
// 	if len(mf.File) <= 0 {
// 		return nil, errors.New("no file uploaded")
// 	}

// 	var header1 *multipart.FileHeader

// 	for _, headers := range mf.File {
// 		header1 = headers[0]
// 		break
// 	}
// 	return header1, nil
// }

// func getUserInGroup(groupID, userID string) (*modeldb.User, error) {

// 	session := mongo_pool.getSession()
// 	if session == nil {
// 		return nil, nil
// 	}
// 	defer mongo_pool.returnSession(session)

// 	var users []*modeldb.User
// 	err := session.DB(mongo_pool.db).C(common.Collection_userinfo).FindId(userID).All(&users)
// 	// err := session.DB(mongo_pool.db).C(common.Collection_userinfo).FindId(bson.M{"group": groupID}).All(&user)
// 	if err != nil {
// 		log.SysF("getUsersFromDB error: %s", err)
// 		return nil, err
// 	}

// 	if users == nil && len(users) <= 0 {
// 		return nil, nil
// 	}
// 	user := users[0]
// 	if user.Group == groupID {
// 		return user, nil
// 	}
// 	return nil, nil
// }

// func getUsersInGroup(groupID string) *models.SafeUserList {
// 	return getUsersFromDB(groupID)

// }

// func getUsersFromDB(groupID string) *models.SafeUserList {
// 	users := models.NewSafeUserList()

// 	session := mongo_pool.getSession()
// 	if session == nil {
// 		return nil
// 	}
// 	defer mongo_pool.returnSession(session)

// 	var users_array modeldb.UserArray
// 	err := session.DB(mongo_pool.db).C(common.Collection_userinfo).Find(bson.M{"group": groupID}).All(&users_array)
// 	if err != nil {
// 		log.SysF("getUsersFromDB error: %s", err)
// 		return nil
// 	}

// 	for _, user := range users_array {
// 		users.Set(user.ID, models.NewUser(models.NewRealUserInfo(user), nil))
// 	}

// 	// session.d.C("")

// 	// res, err := redis_instance.RedisDo("SMEMBERS", common.Set_group_user_id_list+groupID)
// 	// if err != nil {
// 	// 	log.SysF("isUserInGroup error: %s", err)
// 	// 	return nil
// 	// }

// 	// ids, err := redis.Strings(res, nil)
// 	// if err != nil {
// 	// 	log.SysF("isUserInGroup error: %s", err)
// 	// 	return nil
// 	// }

// 	// for _, id := range ids {

// 	// 	res, err := redis_instance.RedisDo("hgetall", common.Hash_users+id)
// 	// 	if err != nil {
// 	// 		log.SysF("getUserInfoFromRedis error: %s", err)
// 	// 		return nil
// 	// 	}
// 	// 	var u modeldb.User
// 	// 	err = redis.ScanStruct(res.([]interface{}), &u)
// 	// 	if err != nil {
// 	// 		log.SysF("getUserInfoFromRedis error: %s", err)
// 	// 		return nil
// 	// 	}

// 	// 	log.TraceF("get user : %#v ", u)
// 	// 	users.Set(u.ID, &u)
// 	// }
// 	return users
// }

// func getUserInfoFromRedis(id string) *modeldb.User {

// 	res, err := redis_instance.RedisDo("hgetall", common.Hash_users+id)
// 	if err != nil {
// 		log.SysF("getUserInfoFromRedis error: %s", err)
// 		return nil
// 	}
// 	var u modeldb.User
// 	err = redis.ScanStruct(res.([]interface{}), &u)
// 	if err != nil {
// 		log.SysF("getUserInfoFromRedis error: %s", err)
// 		return nil
// 	}

// 	log.TraceF("get user : %#v ", u)

// 	return &u
// }
