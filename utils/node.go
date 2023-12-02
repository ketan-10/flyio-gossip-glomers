package utils 

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type Node struct {
  handlers map[string]HandlerFunc
  id string
  nodeIDs []string
  
  mu sync.Mutex
}

type HandlerFunc func(msg Message) error

type Message struct {
  Src string `json:"src,omitempty"`
  Dest string `json:"dest,omitempty"`
  Body json.RawMessage `json:"body,omitempty"`
}

// Message Partial Body. for context of Node, so that we can reply
// MessageBody represents the reserved keys for a message body.
type MessageBody struct {
	// Message type.
	Type string `json:"type,omitempty"`

	// Optional. Message identifier that is unique to the source node.
	MsgID int `json:"msg_id,omitempty"`

	// Optional. For request/response, the msg_id of the request.
	InReplyTo int `json:"in_reply_to,omitempty"`

	// Error code, if an error occurred.
	Code int `json:"code,omitempty"`

	// Error message, if an error occurred.
	Text string `json:"text,omitempty"`
}

type InitMessageBody struct {
  MessageBody
  NodeID string `json:"node_id,omitempty"`
  NodeIDs []string `json:"node_ids,omitempty"`
}

func NewNode() *Node {
  node :=  &Node{
    handlers: make(map[string]HandlerFunc),
  }

  node.Handle("init", func(msg Message) error {

    var body InitMessageBody
    if err := json.Unmarshal(msg.Body, &body); err != nil {
      return fmt.Errorf("unmarshal init message body %w", err)
    }
    node.id = body.NodeID 
    node.nodeIDs = body.NodeIDs

    log.Printf("Node %s initialized", node.id)

    return node.Reply(msg, MessageBody{Type: "init_ok"})

  })

  return node
}

func (n *Node) Handle(typ string, fn HandlerFunc) {
  n.handlers[typ] = fn 
}

func (n *Node) Send(dest string, body any) error {
  bodyJson, err := json.Marshal(body);

  if err != nil {
    return err
  }

  msg, err := json.Marshal(Message{
    Src: n.id,
    Dest: dest,
    Body: bodyJson,
  })

  if err != nil {
    return err
  }

  n.mu.Lock()
  defer n.mu.Unlock()

  log.Printf("Sent %s", msg)

  if _, err = os.Stdout.Write(msg); err != nil {
    return err
  }

  _, err = os.Stdout.Write([]byte{'\n'})
  return err
}

func (n *Node) Reply(req Message, body any) error {
  // Extract message ID form original message
  var reqBody MessageBody
  
  if err := json.Unmarshal(req.Body, &reqBody); err != nil {
    return err
  }

  // marshal then unmarshal body, to inject our reply message id
  b := make(map[string]any)
  buf, err := json.Marshal(body)
  if err != nil {
    return err
  }
  if err := json.Unmarshal(buf, &b); err != nil {
    return err
  }
  b["in_reply_to"] = reqBody.MsgID
  
  return n.Send(req.Src, b)
}

func (n *Node) Run() error {
  scanner := bufio.NewScanner(os.Stdin)

  wg := sync.WaitGroup{}

  for scanner.Scan() {
    line := scanner.Bytes()

    var msg Message
    if err := json.Unmarshal(line, &msg); err != nil {
      return fmt.Errorf("unmarshal message: %w", err)
    }
    
    var body MessageBody
    if err := json.Unmarshal(msg.Body, &body); err != nil {
      return fmt.Errorf("unmarshal body: %w", err)
    }
    log.Printf("Received %s", msg)
    
    var h HandlerFunc
    if h = n.handlers[body.Type]; h == nil {
      // return fmt.Errorf("No hander for %s", line)
      continue; // ingore Message
    }

    wg.Add(1)
    go func() {
      defer wg.Done()
      err := h(msg)
      log.Printf("Exception in handler: %#v:\n%s", msg, err)
    }()
    
  }

  if err := scanner.Err(); err != nil {
    return err
  }

  wg.Wait()

  return nil
}


// type HandlerFunc interface {
//   func send(message Message) error 
//}

