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

#!/usr/bin/env python3

from cProfile import label
import os
import math
import matplotlib.pyplot as plt

PATH     = "results/Protocol"
PREFIX   = "tcp-icmp-shuffled"
TOTAL    = 47000000
STEPSIZE = 1000000

LABELS      = []
PREFIXES    = [i for i in range(0, 129, 4)]
LEVEL_NODES = []
AVG_CHILDREN = []
TOTAL_NODES = []

millnames = ['',' Thousand',' Million',' Billion',' Trillion']

def millify(n):
    n = float(n)
    millidx = max(0,min(len(millnames)-1,
                        int(math.floor(0 if n == 0 else math.log10(abs(n))/3))))

    return '~{:.2f}{}'.format(n / 10**(3 * millidx), millnames[millidx])

for i in range(STEPSIZE, TOTAL + STEPSIZE, STEPSIZE):
    with open(os.path.join(PATH, "{}-{}.txt".format(PREFIX, i))) as fin:
        lines = fin.readlines()
        LEVEL_NODES.append({})
        AVG_CHILDREN.append({})
        for line in lines[:-1]:
            info = line.strip().split("\t")
            prefix = int(float(info[0].split("=")[1]))
            numNodes = int(float(info[1].split("=")[1]))
            avgChildren = float(info[2].split("=")[1])
            maxChildren = int(float(info[4].split("=")[1]))
            maxValue = int(float(info[5].split("=")[1]))
            LEVEL_NODES[-1][prefix] = numNodes
            AVG_CHILDREN[-1][prefix] = avgChildren
        totalNodes = int(lines[-1].strip().split("=")[1])
        LABELS.append(i)
        TOTAL_NODES.append(totalNodes)

# plt.plot(LABELS,TOTAL_NODES, marker='o', label="real")
# plt.title('# of IPs Inserted vs. # of Nodes on Trie')
# plt.xlabel('# of IPs Inserted')
# plt.ylabel('# of Nodes on Trie')
# plt.axline(xy1=(0, 0), xy2=(TOTAL, TOTAL*32), color="black", linestyle=(0, (5, 5)), label="# of IPs vs. (32 * # of IPs)")
# plt.annotate("{}".format(millify(TOTAL_NODES[-1])), # this is the text
#             (LABELS[-1],TOTAL_NODES[-1]), # these are the coordinates to position the label
#             textcoords="offset points", # how to position the text
#             xytext=(0,10), # distance from text to points (x,y)
#             ha='center') # horizontal alignment can be left, right or center
# plt.annotate("{}".format(millify(TOTAL*32)), # this is the text
#             (LABELS[-1],TOTAL*32), # these are the coordinates to position the label
#             textcoords="offset points", # how to position the text
#             xytext=(0,10), # distance from text to points (x,y)
#             ha='center') # horizontal alignment can be left, right or center
# plt.axvline(x=LABELS[-1], ymin=0, color='grey', linestyle='dashed', linewidth=1)
# plt.legend()
# fig = plt.gcf()
# plt.show()
# fig.set_size_inches((15, 15), forward=False)
# fig.savefig("totalNodes-{}.png".format(PREFIX), bbox_inches='tight', dpi=500)

PREV = None
PREV_LABEL = LABELS[0]
for i in range(0, len(LEVEL_NODES)):
    X = PREFIXES
    Y = []
    if PREV != None:
        plt.plot(X,PREV, marker='o', label="{} IPs".format(PREV_LABEL))
        PREV = None
    for prefix in X:
        Y.append(LEVEL_NODES[i][prefix])
    plt.plot(X,Y, marker='o', label="{} IPs".format(LABELS[i]))
    if ((i+1) % 24 == 0):
        plt.title('Prefix (Level) vs. # of Nodes at the Level [{} - {}]IPs'.format(PREV_LABEL, LABELS[i]))
        plt.xlabel('Prefix (Level)')
        plt.ylabel('# of Nodes at the Level')
        plt.legend()
        plt.grid(color='gray', linestyle='--')
        plt.xticks(PREFIXES, PREFIXES)
        fig = plt.gcf()
        plt.show() # show it here (important, if done before you will get blank picture)
        fig.set_size_inches((15, 15), forward=False)
        fig.savefig("[{}-{}]IPs_protocol_filtered.png".format(PREV_LABEL, LABELS[i]), bbox_inches='tight', dpi=500)
        PREV = Y
        PREV_LABEL = LABELS[i]
plt.title('Prefix (Level) vs. # of Nodes at the Level [{} - {}]IPs'.format(PREV_LABEL, LABELS[i]))
plt.xlabel('Prefix (Level)')
plt.ylabel('# of Nodes at the Level')
plt.legend()
plt.grid(color='gray', linestyle='--')
plt.xticks(PREFIXES, PREFIXES)
fig = plt.gcf()
plt.show() # show it here (important, if done before you will get blank picture)
fig.set_size_inches((15, 15), forward=False)
fig.savefig("[{}-{}]IPs_protocol_filtered.png".format(PREV_LABEL, LABELS[i]), bbox_inches='tight', dpi=500)

# PREV = None
# PREV_LABEL = LABELS[0]
# for i in range(0, len(AVG_CHILDREN)):
#     X = PREFIXES
#     Y = []
#     if PREV != None:
#         plt.plot(X,PREV, marker='o', label="{} IPs".format(PREV_LABEL))
#         PREV = None
#     for prefix in X:
#         Y.append(AVG_CHILDREN[i][prefix])
#     plt.plot(X,Y, marker='o', label="{} IPs".format(LABELS[i]))
#     if ((i+1) % 24 == 0):
#         plt.title('Prefix (Level) vs. Avg. # of Children at the Level [{} - {}]IPs'.format(PREV_LABEL, LABELS[i]))
#         plt.xlabel('Prefix (Level)')
#         plt.ylabel('Avg. # of Children at the Level')
#         plt.legend()
#         fig = plt.gcf()
#         plt.show() # show it here (important, if done before you will get blank picture)
#         fig.set_size_inches((8.5, 11), forward=False)
#         fig.savefig("[{}-{}]IPs-avg.png".format(PREV_LABEL, LABELS[i]), bbox_inches='tight', dpi=500)
#         PREV = Y
#         PREV_LABEL = LABELS[i]
# plt.title('Prefix (Level) vs. Avg. # of Children at the Level [{} - {}]IPs'.format(PREV_LABEL, LABELS[i]))
# plt.xlabel('Prefix (Level)')
# plt.ylabel('Avg. # of Children at the Level')
# plt.legend()
# fig = plt.gcf()
# plt.show() # show it here (important, if done before you will get blank picture)
# fig.set_size_inches((8.5, 11), forward=False)
# fig.savefig("[{}-{}]IPs-avg.png".format(PREV_LABEL, LABELS[i]), bbox_inches='tight', dpi=500)

# data = []
# X = PREFIXES
# STEPS = [[x, x+8] for x in range(0, len(LABELS), 8) ]
# STEPS[-1][1] = len(LABELS)
# for step in STEPS:
#     for prefix in X:
#         Y = []
#         for i in range(step[0], step[1]):
#             Y.append(LEVEL_NODES[i][prefix])
#         data.append(Y)
#     plt.boxplot(data, labels=X)
#     plt.title('Prefix (Level) vs. # of Nodes at the Level [{} - {}]IPs'.format(LABELS[step[0]], LABELS[step[1]-1]))
#     plt.xlabel('Prefix (Level)')
#     plt.ylabel('# of Nodes at the Level')
#     fig = plt.gcf()
#     plt.show()
#     fig.set_size_inches((15, 15), forward=False)
#     fig.savefig("[{}-{}]IPs-box.png".format(LABELS[step[0]], LABELS[step[1]-1]), bbox_inches='tight', dpi=500)
#     data = []

# data = []
# X = PREFIXES
# for prefix in X:
#     Y = []
#     for i in range(0, len(LABELS)):
#         Y.append(LEVEL_NODES[i][prefix])
#     data.append(Y)
# plt.boxplot(data, labels=X)
# plt.title('Prefix (Level) vs. # of Nodes at the Level (All Combined without CT)')
# plt.xlabel('Prefix (Level)')
# plt.ylabel('# of Nodes at the Level')
# fig = plt.gcf()
# plt.show()
# fig.set_size_inches((15, 15), forward=False)
# fig.savefig("IPs-box_CT_filtered.png", bbox_inches='tight', dpi=500)
# data = []

# data = []
# X = PREFIXES
# STEPS = [[x, x+8] for x in range(0, len(LABELS), 8) ]
# STEPS[-1][1] = len(LABELS)
# for step in STEPS:
#     for prefix in X:
#         Y = []
#         for i in range(step[0], step[1]):
#             Y.append(AVG_CHILDREN[i][prefix])
#         data.append(Y)
#     plt.boxplot(data, labels=X)
#     plt.title('Prefix (Level) vs. Avg. # of Children at the Level [{} - {}]IPs'.format(LABELS[step[0]], LABELS[step[1]-1]))
#     plt.xlabel('Prefix (Level)')
#     plt.ylabel('Avg. # of Children at the Level')
#     fig = plt.gcf()
#     plt.show()
#     fig.set_size_inches((15, 15), forward=False)
#     fig.savefig("[{}-{}]IPs-box-avg.png".format(LABELS[step[0]], LABELS[step[1]-1]), bbox_inches='tight', dpi=500)
#     data = []

# with open("../20M-shuffled-128.hprof") as f:
#     DATA = eval(f.read())
#     LABELS = [1e5*i for i in range(1, len(DATA) + 1)]
#     # plt.plot(LABELS,DATA, marker='o', label="128-bit")
#     plt.rcParams["figure.autolayout"] = True
#     plt.title('# of IPs Inserted vs. Allocated Memory (MBs) - Based on go\'s runtime library')
#     plt.xlabel('# of IPs Inserted')
#     plt.ylabel('Allocated Memory (MBs)')
#     # default_x_ticks = range(len(LABELS))
#     with open("../20M-shuffled-64.hprof") as f:
#         DATA = eval(f.read())
#         # plt.plot(default_x_ticks,DATA, marker='o', label="top 64-bit")
#         plt.plot(LABELS,DATA, marker='o', label="top 64-bit")
#     plt.legend()
#     fig = plt.gcf()
#     # plt.xticks(default_x_ticks, LABELS, rotation=90)
#     plt.show()
#     fig.set_size_inches((15, 15), forward=False)
#     fig.savefig("MemoryAllocation_64bit.png", bbox_inches='tight', dpi=500)

# with open("../20M-shuffled-128.gprof") as f:
#     DATA = eval(f.read())
#     LABELS = [1e5*i for i in range(1, len(DATA) + 1)]
#     plt.plot(LABELS,DATA, marker='o', label="128-bit")
#     plt.title('# of IPs Inserted vs. Garbage Collector Runtime (s)')
#     plt.xlabel('# of IPs Inserted')
#     plt.ylabel('Garbage Collector Runtime (s)')
#     with open("../20M-shuffled-64.gprof") as f:
#         DATA = eval(f.read())
#         plt.plot(LABELS,DATA, marker='o', label="top 64-bit")
#     plt.legend()
#     fig = plt.gcf()
#     plt.show()
#     fig.set_size_inches((15, 15), forward=False)
#     fig.savefig("GarbageCollector_64bit_vs_128bit.png", bbox_inches='tight', dpi=500)


# with open("50M-intPointer.hprof") as f:
#     DATA = eval(f.read())
#     LABELS = [1e6*i for i in range(1, len(DATA) + 1)]
#     plt.plot(LABELS,DATA, marker='o', label="runtime(htop)")
#     plt.rcParams["figure.autolayout"] = True
#     plt.title('# of IPs Inserted vs. Total Allocated Memory (MBs)')
#     plt.xlabel('# of IPs Inserted')
#     plt.ylabel('Allocated Memory (MBs)')
#     # default_x_ticks = range(len(LABELS))
#     with open("50M-intPointer-gc.hprof") as f:
#         DATA = eval(f.read())
#         # plt.plot(default_x_ticks,DATA, marker='o', label="top 64-bit")
#         plt.plot(LABELS,DATA, marker='o', label="runtime(htop)-with-garbageCollector")
#     with open("50M-intPointer-gc-end.hprof") as f:
#         DATA = eval(f.read())
#         # plt.plot(default_x_ticks,DATA, marker='o', label="top 64-bit")
#         plt.plot(LABELS,DATA, marker='o', label="runtime(htop)-with-garbageCollector-only-at-the-end")
#     with open("allocSpace.pprof") as f:
#         DATA = eval(f.read())
#         # plt.plot(default_x_ticks,DATA, marker='o', label="top 64-bit")
#         plt.plot(LABELS,DATA, marker='o', label="pprof")
#     with open("gcAllocSpace.pprof") as f:
#         DATA = eval(f.read())
#         # plt.plot(default_x_ticks,DATA, marker='o', label="top 64-bit")
#         plt.plot(LABELS,DATA, marker='o', label="pprof-with-garbageCollector")
#     with open("gcEndAllocSpace.pprof") as f:
#         DATA = eval(f.read())
#         # plt.plot(default_x_ticks,DATA, marker='o', label="top 64-bit")
#         plt.plot(LABELS,DATA, marker='o', label="pprof-with-garbageCollector-only-at-the-end")
#     plt.legend()
#     fig = plt.gcf()
#     # plt.xticks(default_x_ticks, LABELS, rotation=90)
#     plt.show()
#     fig.set_size_inches((15, 15), forward=False)
#     fig.savefig("TotalAllocation_realVSgc.png", bbox_inches='tight', dpi=500)

# with open("radixStats/tcp-icmp-shuffled-radix.txt") as f:
#     DATA = eval(f.read())[:20]
#     LABELS = [1e6*i for i in range(1, 21)]
#     plt.plot(LABELS,DATA, marker='o', label="filtered by protocol (tcp80/443 and icmp)")
#     plt.rcParams["figure.autolayout"] = True
#     plt.title('# of IPs Inserted vs # of Nodes: Radix/AMT')
#     plt.xlabel('# of IPs Inserted')
#     plt.ylabel('# of Nodes: Radix/AMT')
#     with open("radixStats/radixStats.txt") as f:
#         DATA = eval(f.read())
#         plt.plot(LABELS, DATA[:len(LABELS)], marker='o', label="with China Telecom")
#     with open("radixStats/radixCTFilteredStats.txt") as f:
#         DATA = eval(f.read())
#         plt.plot(LABELS, DATA[:len(LABELS)], marker='o', label="without China Telecom")
#     # plt.annotate("28,270,205/299,028,376", # this is the text
#     #         (LABELS[-1],DATA[-1]), # these are the coordinates to position the label
#     #         textcoords="offset points", # how to position the text
#     #         xytext=(0,10), # distance from text to points (x,y)
#     #         ha='center') # horizontal alignment can be left, right or center
#     plt.legend()
#     fig = plt.gcf()
#     # plt.xticks(default_x_ticks, LABELS, rotation=90)
#     plt.show()
#     fig.set_size_inches((15, 15), forward=False)
#     fig.savefig("TotalNodesRadixOverAMT_protocolFiltered.png", bbox_inches='tight', dpi=500)