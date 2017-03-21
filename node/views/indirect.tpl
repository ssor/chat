<!DOCTYPE html>
<html lang="en">

<head>
    <title>Chatting </title>
    <link rel="Shortcut Icon" href="/images/favicon.ico">
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
            var proto_reply = 6;
            var proto_close = "loginOnOtherDevice";

            var messageCount = 0
            var needReconnect = true;
            var conn;
            var selfID = "{{.USER}}";
            var groupID = "{{.GROUP}}";
            var msg = $("#msg");
            var uploader;

            var messages = []

            function appendLog(msg) {
                console.log("msg => " + msg)
            }

            // $("#reconnect").click(function() {
            //     conn = null
            //     needReconnect = true
            //     reconnect()
            //     setInterval(reconnect, 5000)
            // });

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
                    messageID: messageCount + ""
                }

                sendData(obj)
                    // return false
            });

            function sendData(obj) {
                conn.send(JSON.stringify(obj));
                messages = _.concat(messages, obj)
                    // appendLog($("<div/>").text("我：" + msg.val()))

                msg.val("");
                console.info("--->>> ", obj)
            }

            // function getID() {
            //     console.info("user: " + selfID + " group: " + groupID)
            //         // createConnection(selfID, groupID)
            //     getLoginInfo(selfID, groupID)
            //     setInterval(reconnect, 5000)
            // }

            // function getLoginInfo(id, group) {
            //     $.getJSON("http://{{.HOST}}/login?id=" + id + "&group=" + group, function(data) {
            //         console.log(data)
            //         if (data.code == 0 && data.data != null) {
            //             var loginInfo = data.data
            //             if (_.size(loginInfo.hosts) > 0 && loginInfo.url.length > 0) {
            //                 var loginUrl = loginInfo.hosts[0] + loginInfo.url
            //                 console.info("try to login with ", loginUrl)
            //                 createConnection(loginUrl)
            //             }
            //         }
            //     })
            // }

            function createConnection(id, group) {
                conn = new WebSocket("ws://{{.HOST}}/ws?id=" + id + "&group=" + group);
                // conn = new WebSocket(url);
                conn.onerror = function(error) {
                    console.error("websocket error: ", error)
                    conn = null
                    set_status_offline()
                }
                conn.onclose = function(evt) {
                    console.info("Connection closed: ", evt)
                    if (evt.reason == proto_close) {
                        // appendLog($("<div><b>连接关闭, 用户在其它终端登录</b></div>"))
                        console.warn("连接关闭, 用户在其它终端登录")
                        needReconnect = false
                        conn = null
                        set_status_offline()
                    } else {
                        // appendLog($("<div><b>连接关闭</b></div>"))
                        conn = null
                        console.warn("连接关闭")
                        set_status_offline()
                    }
                }
                conn.onmessage = function(evt) {
                    console.log(evt.data)
                    var obj = {
                            protocol: proto_reply,
                        }
                        //发送回执
                    sendData(obj)
                    var message = JSON.parse(evt.data)

                    if (message.uid == selfID) {
                        $("#userName").text(message.name)
                    }
                    // if(message.id == selfID && message.protocol != proto_text) return;

                    switch (message.protocol) {
                        case proto_login:
                            if (message.id != selfID) {
                                // appendLog($("<div/>").text(message.name + " 登录"))
                                console.info(message.name + " 登录")
                            }
                            break
                        case proto_logout:
                            if (message.id != selfID) {
                                // appendLog($("<div/>").text(message.name + " 退出"))
                                console.info(message.name + " 退出")
                            }
                            break
                        case proto_text:
                            // appendLog($("<div/>").text(message.name + " ：" + message.content))
                            if (message.uid == selfID) {
                                console.info("<<<--- " + "自己 ：" + message.content)
                            } else {
                                console.info("<<<--- " + message.name + " ：" + message.content)
                            }
                            break
                        case proto_image:
                            // appendLog($("<div>" + message.name + "：<img src='" + message.content + "' style='max-width: 200px;'></div>"))
                            // appendLog($("<div/>").text(message.name + " ：分享了一张图片 "))
                            console.info("<<<--- " + message.name + " ：分享了一张图片 " + message.content)
                            break
                    }
                    // console.log(message)
                    // appendLog($("<div/>").text(evt.data))
                }
                set_status_online()
            }

            // function reconnect() {
            //     if (conn == null && needReconnect == true) {
            //         // createConnection(selfID, groupID)
            //         getLoginInfo(selfID, groupID)
            //     }
            // }

            function initUploader() {

                // 初始化Web Uploader
                uploader = WebUploader.create({

                    // 选完文件后，是否自动上传。
                    auto: true,

                    // swf文件路径
                    swf: 'stylesheets/Uploader.swf',

                    // 文件接收服务端。
                    server: '/uploadImage?uid=' + selfID + "&group=" + groupID,

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
                // getID()
                initUploader()
                createConnection(selfID, groupID)
            } else {
                // appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"))
                alert("Your browser does not support WebSockets")
            }

        });

        function set_status_offline() {
            $("#statusBar").text("已离线...")
        }

        function set_status_online() {
            $("#statusBar").text("交谈中...")
        }
    </script>
    <style type="text/css">
        #form {
            padding: 0 0.5em 0 0.5em;
            margin-left: 15px;
            width: 100%;
            overflow: hidden;
        }
        
        #uploader {
            padding: 0 0.5em 0 0.5em;
            margin-left: 15px;
            width: 100%;
            overflow: hidden;
        }
    </style>
</head>

<body>
    <div id="statusBar" style="font-size: 20px;margin-bottom: 0px;margin-top: 30px;margin-left: 20px;">已离线... </div>

    <h1 id="userName" style="min-height: 50px; margin-left: 20px;"></h1>
    <div id="form">
        <div style="margin-right: 30px;">
            <textarea rows="10" cols="100" id="msg" style="width: 100%"></textarea>

        </div>
        <div style="margin-top: 10px; text-align: right; margin-right: 20px;">
            <input type="button" id="send" value="发送文字" style="width: 10%;font-size: 18px;" />
        </div>
    </div>

    <div id="uploader" class="wu-example" style="margin-top: -25px; width: 50%;">
        <!-- <input type="button" id="imagePicker" value="选择图片" style="text-align:center;width:128px;font-size: 18px;" /> -->
        <div id="imagePicker" style="padding-top:5px;">发送图片</div>
    </div>
    <script type="text/javascript">
    </script>

</body>

</html>