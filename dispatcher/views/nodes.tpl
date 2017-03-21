<!DOCTYPE html>
<html lang="en">

<head>
    <title>节点列表</title>

    <!--引入JS-->
    <script src="javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="javascripts/lodash.js"></script>
    <script type="text/javascript">
        $(function() {

            $("#refresh").click(function() {
                refresh_node_list()
            });

            refresh_node_list()
        });

        // function showNodeGroups(node_host) {
        //     console.log("show node groups ", node_host)
        //     window.location.href = "/groupsIndex?node=" + node_host
        // }

        function refresh_node_list() {
            console.log("refresh_node_list...")
            var node_list = $("#node-list")
            node_list.empty()

            $.get("/nodesInfo", function(res) {
                console.log(res)
                _(res).forEach(function(node_info) {
                    var ele = $('<div> ' + node_info.wan + ' capacity: ' + node_info.capacity + ' current: ' + node_info.current + ' </div>')
                        // ele.click(function() {
                        //     showNodeGroups(node_info.wan)
                        // })
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

    <h1 style="text-align: center;margin-top: 100px; margin-bottom: 40px;">支部在线状态查询</h1>
    <div style="border-bottom: 1px solid rgba(0,0,0,0.2);padding-bottom: 20px; margin-left: 20px;margin-right: 30px;">
        <div style="text-align: center;">
            <input type="text" id="searchContent" value="" style="width: 30%;min-width: 200px;"> </input>
        </div>
        <div style="text-align: center;">
            <input type="button" id="searchAsGroupName" value="搜索该支部" style="margin-top: 10px;width: 10%;font-size: 18px;" />
            <input type="button" id="searchAsUserName" value="搜索该用户" style="margin-top: 10px;width: 10%;font-size: 18px;" />
        </div>
    </div>
    <br>
    <div style="margin-left: 20px;margin-right: 30px;">
        <input type="button" id="refresh" value="刷新" style="margin-top: 10px;width: 10%;font-size: 18px;" />
        <div id="node-list" style="margin-top: 20px;"></div>
    </div>



</body>

</html>