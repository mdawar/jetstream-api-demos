#!/bin/bash

tmux new-session -d -s jetstream-api

tmux split-window -v

tmux select-pane -t 1
tmux split-window -h
tmux split-window -h
tmux select-pane -t 1
tmux split-window -h

tmux select-pane -t 0
tmux split-window -h
tmux split-window -h

tmux send-keys -t 0 'make server' C-m
tmux send-keys -t 1 'sleep 1 && nats-top' C-m
tmux send-keys -t 2 'sleep 1 && make producer' C-m

tmux send-keys -t 3 'sleep 1 && make slowmessagesworker' C-m
tmux send-keys -t 4 'sleep 1 && make slowmessagesworker' C-m
tmux send-keys -t 5 'sleep 1 && make messagesworker' C-m
tmux send-keys -t 6 'sleep 1 && make messagesworker' C-m

tmux attach-session -t jetstream-api
