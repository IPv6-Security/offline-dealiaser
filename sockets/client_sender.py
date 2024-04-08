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
import threading
import subprocess
import json
import time

# Receiver
def create_response_listener():
    r = socket(AF_INET, SOCK_STREAM)
    r.bind(('',6002))
    myaddr, myport = r.getsockname()
    print('bound new socket to {0}:{1}'.format(myaddr, myport))
    r.listen(1)
    conn, remoteaddr = r.accept()
    dar_input = conn.makefile('rb', 0)
    
    # I am not sure if we need this. We can use tee command
    # on the server_sender side as well. Also, the dealiaser
    # has an option to export all the outputs to a file.
    with open('dealiasing_results', 'w+') as file_output:
        proc = subprocess.Popen(
            ['tee'],
            stdin=dar_input,
            stdout=file_output, shell=True
        )
    proc.wait()

    print("Finished")

def run_sending_function():
    background_thread = threading.Thread(target=create_response_listener)
    background_thread.daemon = True
    background_thread.start()

    # Sender
    s = socket(AF_INET, SOCK_STREAM)
    s.connect(('localhost', 6001))
    # Confirm setup first
    data = s.recv(1024).decode()
    #print(data)

    print("Connected")
    dar_output = s.makefile('w',1024)

    # Construct and send commands from IPs
    with open("../../inputs/DealiasingTestIPs.txt", 'r') as g:
        while True:
            line = g.readline()
            if not line:
                break
            data = {}
            data["Type"] = "lookup"
            data["Data"] = line.strip()
            dar_output.write("{}\n".format(json.dumps(data)))
            dar_output.flush()
    
    # When done, send quit command to trigger dealiaser terminate.
    print("Consumed inputs. Sending quit command.")
    data = {}
    data["Type"] = "quit"
    dar_output.write("{}\n".format(json.dumps(data)))
    dar_output.flush()


    #print("Finished Sending")
    dar_output.close()
    s.close()
    #print("Finished Closing Sockets")
    background_thread.join()
    #print("Finished Listening")

    
if __name__ == "__main__":
    run_sending_function()
