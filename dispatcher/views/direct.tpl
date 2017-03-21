<!DOCTYPE html>
<html lang="en">

<head>
    <title>Chat Example</title>
    <!--引入CSS-->
    <link rel="stylesheet" type="text/css" href="stylesheets/webuploader.css">

    <!--引入JS-->
    <script src="javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="javascripts/webuploader.js"></script>
    <script type="text/javascript" src="javascripts/lodash.js"></script>
    <script type="text/javascript">
        $(function() {
            var proto_unknown = -1;
            var proto_login = 0;
            var proto_logout = 1;
            var proto_text = 2;
            var proto_image = 3;
            var proto_close = "loginOnOtherDevice";

            var messageCount = 0
            var conn;
            var needReconnect = true;
            var selfID;
            var groupID;
            var msg = $("#msg");
            var log = $("#log");
            var uploader;

            var messages = []

            function appendLog(msg) {
                var d = log[0]
                var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
                msg.appendTo(log)
                if (doScroll) {
                    d.scrollTop = d.scrollHeight - d.clientHeight;
                }
            }


            $("#reconnect").click(function() {
               conn = null
               needReconnect = true
               reconnect()
               setInterval(reconnect, 5000)
            });

            $("#send").click(function() {
                if (!conn) {
                    return false;
                }
                if (!msg.val()) {
                    return false;
                }
                messageCount++
                var obj = {
                    uid: selfID,
                    protocol: proto_text,
                    content: msg.val(),
                    messageID: "" + messageCount
                }

                conn.send(JSON.stringify(obj));
                messages = _.concat(messages, obj)
                    // appendLog($("<div/>").text("我：" + msg.val()))

                msg.val("");
                return false
            });

            function disp_prompt() {
                var name = prompt("输入用户 ID 和支部 ID, 用 | 隔开, 例如 001|100  ", "")
                if (name != null && name != "") {
                    console.log(name + " login")
                    return name
                }
                return null
            }

            function getID() {
                var id = disp_prompt()

                if (id != null) {
                    var strs = id.split("|")
                    selfID = strs[0]
                    groupID = strs[1]
                    console.log("I am " + selfID)
                    var url = "ws://{{.HOST}}/ws?id=" + selfID + "&group=" + groupID
                    createConnection(url)
                    setInterval(reconnect, 5000)
                }
            }

            function createConnection(url) {
                console.log("createConnection: ", url)
                conn = new WebSocket(url);
                conn.onopen = function(evt) {
                        console.info("websocket open: ", evt)
                    }
                    // conn = new WebSocket("ws://{{.HOST}}/ws?id=" + id + "&group=" + group);
                conn.onerror = function(error) {
                    console.warn("websocket error: ", error)
                    conn = null
                }
                conn.onclose = function(evt) {
                    console.info("Connection closed: ", evt)
                    if (evt.reason == proto_close) {
                        appendLog($("<div><b>连接关闭, 用户在其它终端登录</b></div>"))
                        needReconnect = false
                        conn = null
                    } else {
                        appendLog($("<div><b>连接关闭</b></div>"))
                        conn = null
                    }
                    // setTimeout(reconnect, 5000)
                    // conn = null
                }
                conn.onmessage = function(evt) {
                    console.info("message -> ", evt)
                    var message = JSON.parse(evt.data)

                    if (message.id == selfID) {
                        $("#userName").text(message.name)
                    }
                    // if(message.id == selfID && message.protocol != proto_text) return;

                    switch (message.protocol) {
                        case proto_login:
                            if (message.id != selfID) {
                                appendLog($("<div/>").text(message.name + " 登录"))
                            }
                            break
                        case proto_logout:
                            if (message.id != selfID) {
                                appendLog($("<div/>").text(message.name + " 退出"))
                            }
                            break
                        case proto_text:
                            appendLog($("<div/>").text(message.name + " ：" + message.content))
                            break
                        case proto_image:
                            appendLog($("<div>" + message.name + "：<img src='" + message.content + "' style='max-width: 200px;'></div>"))
                                // appendLog($("<div/>").text(message.name + " ：分享了一张图片 "))
                            break
                    }
                    // console.log(message)
                    // appendLog($("<div/>").text(evt.data))
                }
            }

            function reconnect() {
                if (conn == null && needReconnect == true) {
                    var url = "ws://{{.HOST}}/ws?id=" + selfID + "&group=" + groupID
                    createConnection(url)
                }
            }

            function initUploader() {

                // 初始化Web Uploader
                uploader = WebUploader.create({

                    // 选完文件后，是否自动上传。
                    auto: true,

                    // swf文件路径
                    swf: 'stylesheets/Uploader.swf',

                    // 文件接收服务端。
                    server: '/uploadImage?uid=' + selfID,

                    // 选择文件的按钮。可选。
                    // 内部根据当前运行是创建，可能是input元素，也可能是flash.
                    pick: '#imagePicker',

                    formData: {
                        id: "aaa"
                    },

                    // 只允许选择图片文件。
                    accept: {
                        title: 'Images',
                        extensions: 'gif,jpg,jpeg,bmp,png',
                        mimeTypes: 'image/*'
                    }
                });

                // 文件上传过程中创建进度条实时显示。
                uploader.on('uploadProgress', function(file, percentage) {
                    console.log('process: ', percentage * 100 + '%');
                });

                // 文件上传成功，给item添加成功class, 用样式标记上传成功。
                uploader.on('uploadSuccess', function(file) {
                    console.info("uploadSuccess")
                    console.log(file)
                });

                // 文件上传失败，显示上传出错。
                uploader.on('uploadError', function(file) {
                    console.info('上传失败');
                    console.log(file)
                });

                // 完成上传完了，成功或者失败，先删除进度条。
                uploader.on('uploadComplete', function(file) {
                    console.info("uploadComplete")
                    console.log(file)
                });
            }


            if (window["WebSocket"]) {
                // appendLog($("<div>" + "<img src='/share/images/zsb_1452778107731681648.png' style='max-width: 200px;'></div>"))
                getID()
                initUploader()
            } else {
                appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"))
            }

        });
    </script>
    <style type="text/css">
        html {
            overflow: hidden;
        }
        
        body {
            overflow: hidden;
            padding: 0;
            margin: 0;
            width: 100%;
            height: 100%;
            /*background: gray;*/
        }
        
        #log {
            background: white;
            margin: 0;
            padding: 0.5em 0.5em 0.5em 0.5em;
            position: absolute;
            top: 3.5em;
            left: 0.5em;
            right: 0.5em;
            bottom: 3em;
            overflow: auto;
        }
        
        #form {
            padding: 0 0.5em 0 0.5em;
            margin: 0;
            position: absolute;
            bottom: 50px;
            left: 0px;
            width: 100%;
            overflow: hidden;
        }
        
        #uploader {
            padding: 0 0.5em 0 0.5em;
            margin: 0;
            position: absolute;
            bottom: 0px;
            left: 0px;
            width: 100%;
            overflow: hidden;
        }
    </style>
</head>

<body>
    <input type="button" value="重新连接" id="reconnect" style="width: 30%;font-size: 18px;" />

    <h1 id="userName"></h1>
    <div id="log"></div>
    <div id="form">
        <input type="text" id="msg" style="width: 60%;" />
        <input type="button" id="send" value="发送" style="width: 10%;font-size: 18px;" />

    </div>

    <div id="uploader" class="wu-example">
        <!-- <input type="button" id="imagePicker" value="选择图片" style="text-align:center;width:128px;font-size: 18px;" /> -->
        <div id="imagePicker" style="padding-top:5px;">选择图片</div>
    </div>
    <script type="text/javascript">
    </script>

</body>

</html>