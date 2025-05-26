#!/bin/sh

PROMPT_ID="4bcd57d7-e963-4557-ab22-24fe516712c3"
jq '."4bcd57d7-e963-4557-ab22-24fe516712c3".status.messages | 
    (map(select(.[0] == "execution_start"))[0][1].timestamp) as $start |
    (map(select(.[0] == "execution_success"))[0][1].timestamp) as $end |
    ($end - $start) / 1000' history.log