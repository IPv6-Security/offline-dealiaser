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

# Can be used for memory/CPU usage analysis
topp() (
  time $* &
  pid=(`pgrep -P $!`)
  MAX_CPU=0
  MAX_MEM=0
  AVG_CPU=0
  AVG_MEM=0
  COUNTER=0
  trap ':' INT
  while sleep 0.01
  do 
    CPU=(`ps -o '%cpu=' -p "$pid"`)
    if [[ -z $CPU ]]
    then
      break
    fi
    COUNTER=$((COUNTER + 1))
    AVG_CPU=$((AVG_CPU + CPU))
    if [[ $CPU -gt $MAX_CPU ]]
    then
      MAX_CPU=$CPU
    fi
  done
  # kill "$pid"
  AVG_CPU=$((AVG_CPU / COUNTER))
  echo "AVG_CPU: $AVG_CPU"
  echo "MAX_CPU: $MAX_CPU"
)