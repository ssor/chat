This user package do such things:
    1. receive data from connection, and parse it
    2. handle msg comes in: if a reply, mark the msg send to client successfuly, if not reply, pop to message stroe
	3. receive msg from hub, in other words, from other users, send msg to connection
