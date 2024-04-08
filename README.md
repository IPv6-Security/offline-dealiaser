# Aliasv6
This is the dealiaser component of 6Sense project, which operates on Golang v1.20+ and works on the following principles:

- It works as a lookup table rather than performing actual dealiasing. It inserts a list of given prefixes as a compressed trie, and performs lookups on the constructed radix-like tree. These given prefixes should be known alias prefixes.
- It performs checks on the tree within X intervals to see if there are any updates. If so, it exports the new prefix list as a checkpoint for future use.
- Given two prefixes, if they have the same prefix, but differ on the last bit; dealiaser detects a new aliased prefix under that higher bit (e.g., two /40 prefixes differ at 39th bit -> /39 alias prefix).

## Compiling

In order to compile the code run the following commands:
```
$> go get
$> make
```

## Usage

The tool has several options to control how it works.

```
Usage:
  aliasv6 [OPTIONS]

Application Options:
  -o, --output-file=                   Output filename, use - for stdout (default: -)
  -f, --input-file=                    Input filename, use - for stdin (default: -)
  -m, --metadata-file=                 Metadata filename, use - for stderr (default: -)
  -l, --log-file=                      Log filename, use - for stderr (default: -)
  -c, --construct-input-file=          List of ips/prefixes input file to construct Tree/Trie for tests
      --checkpoint-base-name=          Base name for the Tree/Trie checkpoints if there is a change. It will be followed by the timestamp of the checkpoint (default:
                                       checkpoint)
      --checkpoint-frequency=          Frequency in seconds to export Tree/Trie checkpoints (default: 30.0)
      --test-type=[radix|stats|stress] Testing mode (default: radix)
      --test-input-file=               List of ips input file for Tree/Trie for tests (e.x. lookup tests for radix)
      --test-output-file=              File to export results of stats test mode
  -t, --test                           Set program to testing mode
      --flush                          Flush after each line of output.
      --expanded                       Print IPs in an expanded format
      --test-step-size=                Checkpoint or logging step size for tests (default: 1000000)
      --num-lookup-workers=            Number of workers to perform concurrent lookup operations (default: 1000)
      --input-type=[command|ip]        Input feed type. Command has to be in JSON format, and ip is a IPv6 address as a string. (default: command)

Help Options:
  -h, --help                           Show this help message
```

### Input Type

If the input type if set to `command`, which is default behavior, the program expects a JSON object. There are three available commands.

| Command | Description | Example |
| --- | --- | --- |
| insert | This command performs an insert operation. The data to be inserted have to be a prefix in CIDR format. | `{"Type": "insert", "Data": "ffff:ffff::0000/64"}` |
| lookup | This command performs a lookup operation for the given IP address. If the given data is a prefix, it performs lookup operations for all IP addresses under that prefix range. | `{"Type": "lookup", "Data": "ffff:ffff::1234"}` or `{"Type": "lookup", "Data": "ffff:ffff::0000/96"}` |
| quit | This command terminates the tool, and quits every operation. This might be used by another external tool to send a termination signal to dealiaser. | `{"Type": "quit"}` |

### Testing (Experimental)

The tool also offers testing on the given data if put in testing mode. Testing can be performed in two different data structures, radix or Array Mapped Trie (AMT). Radix uses a compressed trie approach whereas AMT is using a bitmap to optimize the memory usage. Testing can be also done to retrieve statistics about the data or perform stress testing to understand the memory usage.

| Option | Description | Default Value |
| --- | --- | --- |
| `--test` | Set program to testing mode. No value is expected. | |
| `--test-type` | Testing mode. Available options are `radix`, `stats`, `stress`. | `radix` |
| `--test-input-file` | List of IPs/commands input file for Tree/Trie for tests (e.x. lookup tests for radix). | |
| ` --test-output-file` | File to export results of `stats` test mode. Only used in `stats` mode. | |
| `--test-step-size` | Checkpoint or logging step size for tests (number of iterations) | `1000000` |

### Sockets (Experimental)

In order to make `aliasv6` communicate with a different command line tool, one can use the Python scripts present in the `sockets` folder. These scripts implement both client and server
communications (especially, if the communication is handled over a SSH tunnel). However, these scripts are totally experimental, and we do not guarantee that they would work.

### Examples

`./aliasv6 -c prefixes.txt -l aliasv6.log -m aliasv6.meta -o aliasv6.out -f lookupIPs.command`

File `prefixes.txt` should contain a CIDR prefix per line. File `lookupIPs.command` should contain:
```
{"Type": "lookup", "Data": "IP1"}
{"Type": "lookup", "Data": "IP2"}
{"Type": "lookup", "Data": "IP3"}
{"Type": "quit"}
```
Please note that IP1, IP2, IP3 has to be actual IPs in CIDR format. Also, the `quit` command is optional at the end since the file ends with EOF which also triggers termination.

`./aliasv6 -c prefixes.txt -l aliasv6.log -m aliasv6.meta -o aliasv6.out -f lookupIPs.txt --input-type=ip`

Both of the files `prefixes.txt` and `lookupIPs.txt` should contain a CIDR prefix per line. File `lookupIPs.txt` should contain:
```
IP1
IP2
IP3
```
Please note that IP1, IP2, IP3 has to be actual IPs in CIDR format.
