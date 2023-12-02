package main

import (
	"encoding/json"
	"log"
	"os/exec"

	utils "github.com/jepsen-io/maelstrom/demo/go"
)

func main(){

  n := utils.NewNode()

  n.Handle("generate", func(msg utils.Message) error {
    var body map[string]any
    if err := json.Unmarshal(msg.Body, &body); err != nil {
      return err
    }

    newUUID, err := exec.Command("uuidgen").Output()
    if err != nil {
      log.Fatal(err)
    }
  
    body["id"] = newUUID 
    body["type"] = "generate_ok"

    return n.Reply(msg, body)
  }) 

  if err := n.Run(); err != nil {
    log.Fatal(err)
  }
}
