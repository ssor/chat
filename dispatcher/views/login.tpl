<!DOCTYPE html>
<html lang="en">

<head>
    <title>Chat Login</title>

    <!--引入JS-->
    <script src="javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="javascripts/lodash.js"></script>
    <script type="text/javascript">
        $(function() {

            function appendLog(msg) {
                var d = log[0]
                var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
                msg.appendTo(log)
                if (doScroll) {
                    d.scrollTop = d.scrollHeight - d.clientHeight;
                }
            }

            $("#send").click(function() {
                var userID = $("#userID").val()
                var groupID = $("#userGroup").val()

                window.location.href = "/connect?user=" + userID + "&group=" + groupID
            });

        });
    </script>
    <style type="text/css">
        html {
            overflow: hidden;
        }
        
        body {
            overflow: hidden;
            padding: 0;
            margin: 10;
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
            margin: 10px;
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

    <h1>Login</h1>
    <div id="form">
        <div> ID: </div>
        <input type="text" id="userID" style="width: 40%;" />

        <div style="margin-top: 10px;"> </div>

        <div> Group: </div>
        <input type="text" id="userGroup" style="width: 40%;" />
        </br>
        <input type="button" id="send" value="login" style="margin-top: 30px;width: 10%;font-size: 18px;" />

    </div>

    <script type="text/javascript">
    </script>

</body>

</html>