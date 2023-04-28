#!/bin/bash
mkfifo /tmp/f

./projet -id 0 -port 2222 < /tmp/f | ./projet -id 1 -port 3333 | ./projet -id 2 -port 4444 > /tmp/f
