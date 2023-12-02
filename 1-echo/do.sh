go build -o bin
../maelstrom/maelstrom test -w echo --bin bin --node-count 1 --time-limit 10
