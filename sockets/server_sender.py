# Copyright 2024 Georgia Institute of Technology

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# 	http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#! /usr/bin/env python3

from socket import socket, AF_INET, SOCK_STREAM
import sys
import subprocess

# Client Side Sending:
s = socket(AF_INET, SOCK_STREAM)
s.bind(('',6001))
myaddr, myport = s.getsockname()
print('bound new socket to {0}:{1}'.format(myaddr, myport))
s.listen(1)

print("-----------------------------")
conn, remoteaddr = s.accept()
print('accepted connection from {0}:{1}'.format(*remoteaddr))

# Create Connection to Client Server:
r = socket(AF_INET, SOCK_STREAM)
r.connect(('localhost', 6002))
print("Connected to client at: 127.0.0.1:6002")
output_pipe = r.makefile('w', 1024)    
input_pipe = conn.makefile('r',1024)

# Confirm Setup:
conn.send(b'begin')    

print("Begin Subprocess")
commands = ["aliasv6", "-c", "../../inputs/2023_sorted_aliases.txt", "-o", "dealiasing.ljson", "-m", "dealiasing.meta", "-l", "dealiasing.log"]

proc = subprocess.run(
    commands,
    stdin=input_pipe,
    stdout=output_pipe
)

#return_code = proc.wait()

print("Completed Subprocess")
conn.close()
output_pipe.close()
input_pipe.close()
r.close()
