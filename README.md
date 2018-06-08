# QueServer

QueServer is the message queue server which possible to receive messages from several sources
and route them to clients.

## Building

There are no one 3rd parties components.

Place the project into the directory under GOPATH. The directory for the projects:
```
$GOAPTH/queserver
```
It depends on import a local packages **logger**, using to log filtered by levels. 

To build execute a command:
```bash
go build
```

## The Configuration

The following environment variables used to configure the server:

1. **EVENT_SOURCE** - Default: :9090 

    Port or address is listening for event sources. 
    
2. **CLIENT** - Default: :9099
 
    Port or address is listening for event clients.
    
3. **QUEUE_LIMIT** - Default: 1000

    An interval of message sequenceId between a new message and lowest stored in a queue.
    The parameter is like a buffer length for waiting for messages before pulling it in order.
    
4. **QUEUE_TTL** - Default: 500

    Time to live of messages in the queue in milliseconds.
    Messages will pop it from a queue and pull it to send to clients.
     
5. **LOG_LEVEL** - Default: Info

    Set as "Debug" to show detailed logs.
    
## Example running

Default:
```bash
./queserver
```

Changed queue parameters:
```bash
QUEUE_LIMIT=100 QUEUE_TTL=1000 ./queserver
```

Press Ctrl+C to stop the server.

## Architecture

The server contains following parts, ordered by message routing:

1. **EventSource** - _eventSource.go_

    Listener of the EVENT_SOURCE port to receive a message from several event sources.
    The message payload will convert to internal structure.
    Then the messages will push to queue.
    
2. **Queue** - _queue.go_

    A queue is a buffer for messages.
    It accumulates messages to pull it next in sequence id order.
    Messages will pull next by **QUEUE_LIMIT** threshold or after out of **QUEUE_TTL**.
    Queue skips invalid messages (empty sequence id or unknown message type).

3. **Router** - router.go

    Routing messages by type to the clients and stores followers information.
    Messages will send to the registered client and no action for unregistered users.
    
4. **Server** - server.go

    Listener of the CLIENT port to register clients in the router.
    New clients will process messages in a separated goroutines.
    
5. **Client** - client.go

    The client connection is using to send messages to the registered client.
    
6. **Statistics** - statistics.go

    Collecting receiving/sending statistics of processing messages.