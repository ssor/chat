# Chat

Chat is a websocket based chat system. It's simple but powerful. It can let millions of people chat on line at the same time.


## Design
The whole chat system includes two parts: dispatcher and node. 

### node
node is responsible for:

    1. create connection with clients
    2. handle text message, audio message, and image share message
    3. broadcast message from client to other clients in same group

### dispatcher
dispatcher is responsible for:

    1. load data to redis cache from mongo
    2. handle nsq events when data in mongo changed, and refresh data in cache
    3. check nodes status, add new node, remove dead node
    4. dispatch group to node, and balance groups on nodes
    5. handle clients requests for node info for logging



## Requirement

### Mongo
Stores data used in system

### Redis
Cache tool 

### Nsq
Message queue, notify dispatcher to refresh data

## Build

1. build node

    ``` python ./node/build_release.py```

2. build dispatcher

    ``` python ./dispatcher/build_release.py```

If build successfully, binary file will be in dir:  

> - ./node/release/
> - ./dispatcher/release


## Run

### 1. run mongo

### 2. run redis

### 3. run nsq

### 4. run dispatcher

### 5. run node
