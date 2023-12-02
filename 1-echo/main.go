package main

import (
	"encoding/json"
	"log"

	utils "github.com/jepsen-io/maelstrom/demo/go"
)

func main(){

  n := utils.NewNode()

  n.Handle("echo", func(msg utils.Message) error {
    var body map[string]any
    if err := json.Unmarshal(msg.Body, &body); err != nil {
      return err
    }
    
    body["type"] = "echo_ok"

    return n.Reply(msg, body)
  }) 

  if err := n.Run(); err != nil {
    log.Fatal(err)
  }
}
