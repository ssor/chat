<!DOCTYPE html>
<html lang="en">

<head>
    <title>Group List On Node</title>

    <!--引入JS-->
    <script src="javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="javascripts/lodash.js"></script>
    <script type="text/javascript">
        $(function() {

            $("#refresh").click(function() {
                refresh_status()
            });

            refresh_status()
        });

        function refresh_status() {
            console.log("refresh_status...")
            var node_summary = $("#node-summary")
            node_summary.empty()
            var node_list = $("#node-list")
            node_list.empty()

            $.get("/status", function(res) {
                console.log(res)
                node_summary.append("<div>总人数:   " + res.userTotal + "</div>")
                node_summary.append("<div>在线人数: " + res.userOnline + "</div>")

                _(res.reports).forEach(function(report) {
                    var ele = $('<div> ID: ' + report.id + ' total: ' + report.userTotal + ' online: ' + report.userOnline + ' </div>')
                    node_list.append(ele)
                })
            })
        }
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

    <h1>节点状态</h1>
    <input type="button" id="refresh" value="刷新" style="margin-top: 10px;width: 10%;font-size: 18px;" />
    <div id="node-summary" style="margin-top: 20px;"> </div>
    <div id="node-list" style="margin-top: 20px;"> </div>



</body>

</html>