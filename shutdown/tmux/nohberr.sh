#!/bin/bash

tmux new-session -d -s jetstream-api

tmux split-window -v
tmux select-pane -t 0
tmux split-window -h

tmux send-keys -t 0 'make server' C-m
tmux send-keys -t 1 'sleep 1 && nats-top' C-m
tmux send-keys -t 2 'sleep 1 && go run ./cmd/messages/main.go --hberr=false' C-m

tmux select-pane -t 2

tmux attach-session -t jetstream-api
