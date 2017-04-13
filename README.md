# Chat

Chat is a websocket based chat system. It's simple but powerful. It can let millions of people chat on line at the same time.


## Design
The whole chat system includes two parts: dispatcher and node. 

### dispatcher

### node

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
