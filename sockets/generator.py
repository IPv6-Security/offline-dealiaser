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

from subprocess import Popen, PIPE
import json
import time

def create_command(type, data=None):
    if type not in ["lookup", "insert", "quit"]:
        return {}
    command = {"Type": type}
    if type in ["lookup", "insert"]:
        if not data:
            return {}
        command["Data"] = data
    return command

def receiver(queue):
    pass

def main():
    commands = ["aliasv6", "--flush", "-c", "../../inputs/2023_sorted_aliases.txt", "-l", "dealiaser.log", "-m", "dealiaser.meta"]
    process = Popen(commands, stdin=PIPE, stdout=PIPE)
    tee_commands = ['tee', 'dealiaser.results']
    tee_process = Popen(tee_commands, stdin=process.stdout, stdout=PIPE)
    time.sleep(1)

    dealised_IPs = []
    reader_counter = 0
    last_count = 0
    st = time.time()

    with open("../../inputs/DealiasingTestIPs.txt", 'r') as g:
        while True:
            line = g.readline()
            if not line:
                process.stdin.flush()
                break
            reader_counter += 1
            line = line.strip()
            process.stdin.write(bytes(f"{line}\n", encoding="utf-8"))
            
            # data = create_command("lookup", line.strip())
            # if not data:
            #     continue
            # process.stdin.write(bytes("{}\n".format(json.dumps(data)), encoding='utf-8'))
            # process.stdin.flush()
            
            if reader_counter % 1000 == 0:
                process.stdin.flush()
            et = time.time()
            # print(et - st, reader_counter)
            if et - st >= 1.0:
                print(et - st, reader_counter - last_count)
                last_count = reader_counter
                st = et
            if reader_counter % 1000 == 0:
                # print("reading")
                for _ in range(0, 1000):
                    results = json.loads(tee_process.stdout.readline())
                    ip = results["ip"] # the ip we sent
                    status = results["status"] # "no-match" or "success"
                    timestamp = results["timestamp"]
                    res = results["result"]["aliased"] # true or false
                    if res and "metadata" in results["result"]:
                        alias_prefix = results["result"]["metadata"]
                    else:
                        dealised_IPs.append(ip)
    for _ in range(0, reader_counter % 1000):
        results = json.loads(tee_process.stdout.readline())
        ip = results["ip"] # the ip we sent
        status = results["status"] # "no-match" or "success"
        timestamp = results["timestamp"]
        res = results["result"]["aliased"] # true or false
        if res and "metadata" in results["result"]:
            alias_prefix = results["result"]["metadata"]
        else:
            dealised_IPs.append(ip)

    # data = create_command("insert", "ffff:ffff:ffff:ffff::/64")
    # if not data:
    #     print("something is wrong")
    # process.stdin.write(bytes("{}\n".format(json.dumps(data)), encoding='utf-8'))
    # process.stdin.flush()

    # process.stdin.write(bytes(f"ffff:ffff:ffff:ffff:dead:beef:0101:0234\n", encoding="utf-8"))
    # process.stdin.flush()

    # data = create_command("quit")
    # if not data:
    #     print("something is wrong")
    # process.stdin.write(bytes("{}\n".format(json.dumps(data)), encoding='utf-8'))
    # process.stdin.flush()

    et = time.time()
    print("finished", et - st, reader_counter - last_count)

    process.stdin.flush()
    process.stdin.close()
    process.stdout.close()
    print('Waiting for tee to exit')
    tee_process.wait()
    print('tee finished with return code %d' % tee_process.returncode)
    
if __name__ == "__main__":
    main()