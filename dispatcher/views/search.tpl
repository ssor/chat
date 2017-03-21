<!DOCTYPE html>
<html lang="en">

<head>
    <title>搜索</title>
    <link rel="Shortcut Icon" href="/images/favicon.ico">
    <!--引入JS-->
    <script src="javascripts/jquery.min.js"></script>
    <script type="text/javascript" src="javascripts/lodash.js"></script>
    <link rel="stylesheet" href="/datatable/css/jquery.dataTables.min.css">
    <script type="text/javascript" src="/datatable/js/jquery.dataTables.min.js"></script>
    <script type="text/javascript">
        var table = null;

        $(function() {
            refresh_node_list()

            table = optTable()

        });

        function join_chatroom() {
            var selected_row = get_selected_row()
            if (selected_row == null) {
                console.warn("no group selected")
            }
            // console.log(selected_row)
            var data = selected_row.data()
            console.log("join chatroom " + data[2] + "...")
            goto_group(data[3], data[1])
        }

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
                node_list.text("当前节点数量: " + _.size(res))

            })
        }

        function search_group_name() {
            var content = $("#searchContent").val()
            console.log("search => " + content)
            $.get("/search?group=" + content, function(res) {
                console.log(res)
                table.row().remove().draw(false);
                // if (table != null) {
                //     table.destroy();
                //     table = optTable()
                // }
                _(res).forEach(function(group, index) {
                    table.row.add([index + 1, group.id, group.name, group.node]) //.draw(false)
                })
                table.draw(false)
            })
        }

        function goto_group(node, group) {
            var url = "http://" + node + "?group=" + group
            console.log("go to group ", url)
            window.open(url)
        }

        function get_selected_row() {
            var row = table.row('.selected')
            return row
        }

        function optTable() {
            var table = $('#table').DataTable({
                "oLanguage": {
                    "sLengthMenu": "每页显示 _MENU_ 条记录",
                    "sInfo": "从 _START_ 到 _END_ /共 _TOTAL_ 条数据",
                    "sInfoEmpty": "没有数据",
                    "sInfoFiltered": "(从 _MAX_ 条数据中检索)",
                    "oPaginate": {
                        "sFirst": "首页",
                        "sPrevious": "前一页",
                        "sNext": "后一页",
                        "sLast": "尾页"
                    },
                    "sZeroRecords": "没有检索到数据"
                },
                "paging": false,
                "ordering": false,
                "info": false
            });

            $('#table tbody').on('click', 'tr', function() {
                if ($(this).hasClass('selected')) {
                    $(this).removeClass('selected');
                } else {
                    table.$('tr.selected').removeClass('selected');
                    $(this).addClass('selected');
                }
            });

            return table;
        }
    </script>

</head>

<body>

    <h1 style="text-align: center;margin-top: 100px; margin-bottom: 40px;">支部在线状态查询</h1>
    <div style="border-bottom: 1px solid rgba(0,0,0,0.2);padding-bottom: 20px; margin-left: 20px;margin-right: 30px;">
        <div style="text-align: center;">
            <input type="text" id="searchContent" value="25_16-01-50" style="width: 30%;min-width: 200px;"> </input>
        </div>
        <div style="text-align: center;margin-top: 30px; margin-bottom: 80px;">
            <input type="button" id="searchAsGroupName" value="搜索该支部" style="width: 10%;font-size: 18px;" onclick="search_group_name()" />
            <input type="button" id="searchAsUserName" value="搜索该用户" style="width: 10%;font-size: 18px;" />
        </div>
        <div style="height: 50px;">
            <input type="button" id="joinChatroom" value="加入支部聊天" style="width: 10%;font-size: 18px;margin-top: 29px;" onclick="join_chatroom()" />
        </div>
        <div style="text-align: center;">
            <table id="table" class="display table" cellspacing="0">
                <thead>
                    <tr>
                        <th></th>
                        <th>ID</th>
                        <th>支部名称</th>
                        <th>所在节点</th>
                    </tr>
                </thead>
                <tbody>
                </tbody>
            </table>
        </div>
    </div>

    <br>
    <div style="margin-left: 20px;margin-right: 30px;">
        <div id="node-list" style="text-align: center;font-size: 13px;"></div>
    </div>

</body>

</html>