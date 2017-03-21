<!DOCTYPE html>
<html lang="en">

<head>
    <title>Chat Login</title>
    <link rel="Shortcut Icon" href="/images/favicon.ico">

    <!--引入JS-->
    <script src="javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="javascripts/lodash.js"></script>
    <script type="text/javascript">
        $(function() {

            $("#send").click(function() {
                var userID = $("#userID").val()
                var groupID = $("#userGroup").val()

                window.location.href = "/connect?user=" + userID + "&group=" + groupID
            });

            $("#userfake").click(function() {
                $("#userID").val("iamafakeuser")
            })
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

    <h1>登录</h1>
    <div id="form">
        <div> ID: </div>
        <input type="text" id="userID" style="width: 40%;" />
        <input type="button" id="userfake" value="虚拟用户" style="font-size: 18px;" />

        <div style="margin-top: 10px;"> </div>

        <div> Group: </div>
        <input type="text" id="userGroup" style="width: 40%;" value="{{.GROUP}}" />
        </br>
        <input type="button" id="send" value="login" style="margin-top: 30px;width: 10%;font-size: 18px;" />

    </div>
    <!--<a href="/statusIndex" target="_blank">status show</a>-->
    <script type="text/javascript">
    </script>

</body>

</html>