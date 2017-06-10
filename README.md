# bcr 
Broad Chat Game Project

### Compile with debug info

```bash
make GOFLags="-gcflags -N"
```

## What is Plan ?

```monodraw
                                        ┌─────────────┐
                                        │             │
                                        │   clients   │
                                        │             │
                                        └─────────────┘
                                               │
                                               │
                                               │
                                               │
                                ┌──────────────┴────────────┐
                                │                           │
                                │                           │
                                │                           │
                                │                           │
                                ▼                           ▼
                         ┌─────────────┐             ┌─────────────┐
                         │             │             │             │
                         │   gateway   │             │  websocket  │
                         │             │             │             │
                         └─────────────┘             └─────────────┘
                                │                           │
                                │                           │
                                │                           │
                                │                           │
                                │                           │
       ┌────────────────────────┴───────────────────────────┴─────────────────────────┐
       │                                                                              │
       │                                                                              │
       │                                                                              │
       │                                                                              │
       ▼                                                                              ▼
┌─────────────┐                                                                ┌─────────────┐
│             │                                                                │             │
│    user     │                                                                │    other    │
│             │                                                                │             │
└─────────────┘                                                                └─────────────┘
```

### specify broker or registry services address before start service

```bash
./bin/{YOU-SERVICE} --broker=nats --broker_address=127.0.0.1:4222 --registry=consul --registry_address=127.0.0.1:8500
```

### Default Config
```json
{
    "log_level": "debug",
    "db_url": "root:@tcp(127.0.0.1:3306)/bcr?parseTime=true",
    "redis_host": "127.0.0.1:6379",
    "api_host": "0.0.0.0:8080",
    "web_host": "0.0.0.0:8999"
}
```

### Init Admin User

```bash
    make tool
    ./bin/admin --username="adminuser" --password="adminuser"
```
