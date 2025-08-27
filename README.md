# ZKP Trusted Setup Ceremony Coordinator

<!-- TOC -->
* [ZKP Trusted Setup Ceremony Coordinator](#zkp-trusted-setup-ceremony-coordinator)
  * [Overview](#overview)
    * [Online mode](#online-mode)
    * [Offline mode](#offline-mode)
    * [Snarkjs powers of tau (ptau) -> Phase 1 conversion](#snarkjs-powers-of-tau-ptau---phase-1-conversion)
  * [Prerequisites](#prerequisites)
  * [Build](#build)
  * [Usage](#usage)
  * [Commands](#commands)
    * [General purpose commands](#general-purpose-commands)
      * [`help`](#help)
      * [`ptau`](#ptau)
    * [Online mode commands](#online-mode-commands)
      * [`server`](#server)
        * [Using AWS S3 for artifacts storage](#using-aws-s3-for-artifacts-storage)
      * [`client`](#client)
    * [Offline mode commands](#offline-mode-commands)
      * [`init`](#init)
      * [`contribute`](#contribute)
      * [`verify`](#verify)
      * [`extract-keys`](#extract-keys)
<!-- TOC -->

## Overview
This utility program allows for performing a Trusted Setup Ceremony in a Multi-Party Computation fashion. It is meant
to be used by the Coordinator of the ceremony, as well as by the Contributors. In the end, the Coordinator will obtain
Proving and Verifying Keys, which can be used to generate proofs for the circuit the ceremony was conducted for.

**This project is designed to work with ZK circuits implemented in [Gnark](https://github.com/consensys/gnark).**

Gnark version used for implementing the circuit the ceremony will be conducted for must match the Gnark version used
in this project. Please consult [`go.mod`](./go.mod) to learn which version of Gnark is used. Make sure your circuit
uses the same Gnark version.

Your Gnark project must also satisfy the following constraints:
- Supported curve: BN254
- Supported backend: Groth16


### Online mode

The primary mode of the program. In this mode, the Coordinator runs the ceremony server, which is responsible for
accepting contributions from the Contributors. The Contributors connect to the Coordinator and contribute to the
ceremony.

See help for [`server`](#server) and [`client`](#client) commands for details.

### Offline mode

In this mode, the Coordinator and the Contributors run the ceremony locally. The Coordinator initializes the ceremony
and generates the initial Phase 2 file. The Coordinator sends the file to the first Contributor. The Contributor
generates their contribution and sends them to the Coordinator in the form of a Phase 2 file. The Coordinator verifies
the contributions and, if the verification is positive, sends it to the next Contributor.

In this mode, sending Phase 2 files must be performed manually by the Coordinator and Contributors.

At the end of the ceremony, the Coordinator will have a list of accepted contributions. The Coordinator can then
perform the final verification and extract the Proving and the Verifying Keys.

See help for [`init`](#init), [`contribute`](#contribute), [`verify`](#verify) and [`extract-keys`](#extract-keys)
commands for details.

### Snarkjs powers of tau (ptau) -> Phase 1 conversion

The tool can convert a Snarkjs powers of tau file to a Phase 1 file. This step is performed by the Coordinator before
the initialization of the offline mode ceremony, if the Coordinator has a ptau file that they wish to use in the ceremony.

This step is not necessary if the Coordinator already has a Phase 1 file.

## Prerequisites

These are one-time steps that must be done in order to build the program.

Install [Go](https://go.dev/dl/). Any recent version will do. Look into [`go.mod`](./go.mod) to see the minimum required version.

Install [Protocol Buffer Compiler](https://protobuf.dev/installation/).

Install gRPC for Go:

```shell
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Build

To build the project, run:
```shell
$ go generate ./...
$ go build .
````

in the project's root directory.

The test suite can be executed with:
```shell
$ go test -v ./...
```

## Usage

Run the program with:
```shell
$ go run . <command> <options>

# or, after the program was built
$ ./trusted-setup <command> <options>
```

Running the program with no arguments lists the available commands. Running the program with the command but without
options will display the command's help.

## Commands

### General purpose commands

#### `help`

Print help.

#### `ptau`

Convert a Snarkjs powers of tau file to a Phase 1 file. This step is performed by the Coordinator.

The tool does not initialize its own Phase 1 of the ceremony. It is expected that the Coordinator will generate the
Phase 1 file themselves. As a convenience utility, this command allows the Coordinator to convert a Snarkjs powers of
tau file to a Phase 1 file, which can be used to initialize the Phase 2 of the ceremony.

- `--ptau` - A Snarkjs powers of tau file,
- `--phase1` - The output Phase 1 file.

Example usage:

```shell
$ ./trusted-setup ptau --ptau test.ptau --phase1 test.ph1
2025/09/02 00:31:32 Convert Starkjs powers of tau to Phase 1:
        Load ptau from:   offline/test/resources/test.ptau
        Store Phase 1 to: test.ph1
2025/09/02 00:31:32 Loading Starkjs powers of tau from offline/test/resources/test.ptau
// ... some details of the ptau file, skipped for brevity ...
2025/09/02 00:31:32 Converting Starkjs powers of tau to Phase 1
2025/09/02 00:31:32 Storing Phase 1 to test.ph1
2025/09/02 00:31:32 Operation successful
```

### Online mode commands

#### `server`

Start a Ceremony server. This step is performed by the Coordinator.

The server is responsible for orchestrating the ceremony, receiving contributions from the [participants](#client) and,
in the end, generating Proving and Verifying Keys.

The server is configured with a JSON file. An example configuration is shown below:
```json5
{
  // A human-readable name for the ceremony that will be sent to contributors.
  // Used for identification purposes; can be any reasonably sized string.
  "ceremonyName": "test ceremony",
  // The IP address on which the server will listen on.
  "host": "127.0.0.1",
  // The TCP port on which the server will listen on.
  "port": 7312,
  // The path to the R1CS file generated from a Gnark circuit.
  "r1cs": "resources/server.r1cs",
  // The path to the Phase 1 file (possibly generated from a ptau file - see the `ptau` command for details).
  "phase1": "resources/server.ph1",
  // (optional) If true, AWS S3 is chosen as the storage backend. If false or not present,
  // the server will fall back to storing artifacts in tmpfs.
  "useS3": true,
  // (optional) Name of the AWS S3 bucket to store ceremony artifacts. The bucket must exist.
  // If not provided, this information is taken from a default source (env or AWS CLI config file).
  "s3Bucket": "my-ceremony-bucket",
  // (optional) Region of the AWS S3 bucket to store ceremony artifacts.
  // If not provided, this information is taken from a default source (env or AWS CLI config file).
  "s3Region": "us-east-1",
  // (optional) Profile of the AWS S3 bucket to store ceremony artifacts.
  // If not provided, this information is taken from a default source (env or AWS CLI config file).
  "s3Profile": "test",
  // (optional) Credentials file for AWS S3.
  // If not provided, this information is taken from a default source (env or AWS CLI config file).
  "s3CredentialsFile": "~/.aws/credentials",
}
```

Coordination of the ceremony is automatic. No action from the Coordinator is required besides starting the server
and stopping it with CTRL+C at any arbitrary moment. At CTRL+C, the server stops accepting new contributions and starts
key extraction from the existing contributions.

At the end of the contribution, the artifacts are either saved in the temporary filesystem or in the AWS S3, depending
on the provided configuration. The stored artifacts are:
- SRS commons of the circuit,
- all intermediate Phase 2 files,
- Proving Key,
- Verifying Key,
- JSON structured log documenting the ceremony.

These artifacts are enough to re-generate the keys again using the offline mode [`extract-keys`](#extract-keys) command.

- `--config` - Path to a JSON file containing the server configuration.

Example usage:

```shell
$ ./trusted-setup server --config small.json
2025/09/02 00:33:00 Loading config file: small.json
2025/09/02 00:33:00 Loading R1CS from online/test/resources/server.r1cs
2025/09/02 00:33:00 Loading Phase 1 from online/test/resources/server.ph1
2025/09/02 00:33:01 INF beacon=a40d1701974c7e804732d3bf21131137f6b05dc054b073d0e48110ed2099f11c
2025/09/02 00:33:01 Ceremony artifacts will be stored in tmpfs
2025/09/02 00:33:01 Initializing Phase 2
2025/09/02 00:33:02 INF new ceremony started name=testCeremony
2025/09/02 00:33:02 Server started, waiting for Contributors on 127.0.0.1:7312...
2025/09/02 00:33:02 Press Ctrl+C to end Ceremony and generate Keys
2025/09/02 00:33:09 INF new contributor connected ip=127.0.0.1:64624
2025/09/02 00:33:09 INF contributor position update ip=127.0.0.1:64624 newQueuePosition=0
2025/09/02 00:33:09 INF sending last accepted contribution ip=127.0.0.1:64624
2025/09/02 00:33:09 INF sent last accepted contribution ip=127.0.0.1:64624 size=27147
2025/09/02 00:33:09 INF receiving new contribution candidate ip=127.0.0.1:64624
2025/09/02 00:33:09 INF new contribution candidate accepted ip=127.0.0.1:64624 size=27179
^C2025/09/02 00:33:15 Generating keys out of 1 contributions...
2025/09/02 00:33:15 Artifacts generated in the ceremony:
2025/09/02 00:33:15     /var/folders/jh/bj77hlmj3k50qyxt3tzlbrwc0000gn/T/testCeremony-srs-commons-1639893806
2025/09/02 00:33:15     /var/folders/jh/bj77hlmj3k50qyxt3tzlbrwc0000gn/T/testCeremony-phase2-1-3080991318
2025/09/02 00:33:15     /var/folders/jh/bj77hlmj3k50qyxt3tzlbrwc0000gn/T/testCeremony-pk-2115635955
2025/09/02 00:33:15     /var/folders/jh/bj77hlmj3k50qyxt3tzlbrwc0000gn/T/testCeremony-vk-2427354540
2025/09/02 00:33:15     /var/folders/jh/bj77hlmj3k50qyxt3tzlbrwc0000gn/T/testCeremony-log-454330672
2025/09/02 00:33:15 Operation successful
```

Example structured JSON log from the above ceremony:

```json
{"level":"info","beacon":"a40d1701974c7e804732d3bf21131137f6b05dc054b073d0e48110ed2099f11c","time":"2025-09-02T00:33:01+02:00"}
{"level":"info","name":"testCeremony","time":"2025-09-02T00:33:02+02:00","message":"new ceremony started"}
{"level":"info","ip":"127.0.0.1:64624","time":"2025-09-02T00:33:09+02:00","message":"new contributor connected"}
{"level":"info","newQueuePosition":0,"ip":"127.0.0.1:64624","time":"2025-09-02T00:33:09+02:00","message":"contributor position update"}
{"level":"info","ip":"127.0.0.1:64624","time":"2025-09-02T00:33:09+02:00","message":"sending last accepted contribution"}
{"level":"info","ip":"127.0.0.1:64624","size":27147,"time":"2025-09-02T00:33:09+02:00","message":"sent last accepted contribution"}
{"level":"info","ip":"127.0.0.1:64624","time":"2025-09-02T00:33:09+02:00","message":"receiving new contribution candidate"}
{"level":"info","ip":"127.0.0.1:64624","size":27179,"time":"2025-09-02T00:33:09+02:00","message":"new contribution candidate accepted"}
```

##### Using AWS S3 for artifacts storage

The server can store the ceremony artifacts in AWS S3.

AWS S3 is chosen as the storage backend if `useS3` configuration field is set to `true`. If the field is `false` or not
present, the server will use tmpfs for storage. Some S3 settings can be overridden. See [`server`](#server)
for the details on the configuration file.

AWS S3 credentials are loaded from the credentials file. This file is automatically generated by the
[AWS CLI tool](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html). The file can be
generated as follows:

```shell
$ aws configure --profile <profile name>
AWS Access Key ID [None]: <ID>
AWS Secret Access Key [None]: <key>
Default region name [None]: <region name>
Default output format [None]:
```

After the ceremony, the bucket can be queried as follows:

```shell
$ AWS_PROFILE=test aws s3 ls s3://ceremony-bucket
2025-09-02 01:58:10        998 log
2025-09-02 01:58:05      27179 phase2-1
2025-09-02 01:58:09      59955 pk
2025-09-02 01:58:04      98248 srs-commons
2025-09-02 01:58:10        364 vk
```

Artifacts can be downloaded as follows:

```shell
$ AWS_PROFILE=test aws s3 cp s3://ceremony-bucket/ ceremony_artifacts --recursive
download: s3://ceremony-bucket/log to ceremony_artifacts/log
download: s3://ceremony-bucket/vk to ceremony_artifacts/vk
download: s3://ceremony-bucket/phase2-1 to ceremony_artifacts/phase2-1
download: s3://ceremony-bucket/pk to ceremony_artifacts/pk
download: s3://ceremony-bucket/srs-commons to ceremony_artifacts/srs-commons

$ ls -l ceremony_artifacts
total 384
-rw-r--r--@ 1 user  group    998 Sep  2 01:58 log
-rw-r--r--@ 1 user  group  27179 Sep  2 01:58 phase2-1
-rw-r--r--@ 1 user  group  59955 Sep  2 01:58 pk
-rw-r--r--@ 1 user  group  98248 Sep  2 01:58 srs-commons
-rw-r--r--@ 1 user  group    364 Sep  2 01:58 vk
```

#### `client`

Connect to a Ceremony [server](#server) and provide contributions. This step is performed by the Contributors.

The client is responsible for connecting to the server and providing contributions. The client is configured with
a host and port of the server. Participation in the ceremony is automatic. No action from the Contributor is required
besides starting the client.

- `--host` - The IP address of the server,
- `--port` - The port of the server.

Example usage:

```shell
$ ./trusted-setup client --host 127.0.0.1 --port 7312
2025/09/02 00:33:09 Connecting to 127.0.0.1:7312...
2025/09/02 00:33:09 Joined ceremony: testCeremony
2025/09/02 00:33:09 Contribution slot assigned, position in queue: 0
2025/09/02 00:33:09 Our turn, downloading last contribution
2025/09/02 00:33:09 Received 27147 bytes
2025/09/02 00:33:09 Generating contribution
2025/09/02 00:33:09 Uploading our contribution
2025/09/02 00:33:09 Sent 27179 bytes
2025/09/02 00:33:09 Contribution accepted
2025/09/02 00:33:09 Operation successful
```

### Offline mode commands

#### `init`

Initialize Phase 2 of the ceremony for the given R1CS with a Phase 1 file. This step is performed by the Coordinator.

This step outputs a Phase 2 file based on the provided R1CS and Phase 1 file. The Coordinator must provide the R1CS file
generated from a Gnark circuit and the Phase 1 file either generated in the previous step or from another
cryptographically safe source.

The R1CS file can be generated by the Gnark project implementing the circuit the ceremony will be held for. Please
see the [Constraints](#Constraints) section for more information.

The output Phase 2 file can be used for later contributions.

The command outputs a beacon value, which must then be passed as an argument to [`extract-keys`](#extract-keys).

- `--r1cs` - The R1CS file generated from a Gnark circuit,
- `--phase1` - The Phase 1 file,
- `--phase2` - The output path for the Phase 2 file,
- `--srscommons` - The output path for circuit-independent components of the Groth16 SRS.

Example usage:

```shell
$ ./trusted-setup --phase1 online/test/resources/server.ph1 --r1cs online/test/resources/server.r1cs --phase2 tiny.ph2 --srscommons tiny.srs
2025/09/02 00:39:21 Initializing Phase 2:
        Load Phase 1 from:                online/test/resources/server.ph1
        Load R1CS from:                   online/test/resources/server.r1cs
        Store Phase 2 to:                 tiny.ph2
        Store SRS commons to:             tiny.srs
        Beacon (pass it to extract-keys): b570d6c3cb53603fc4d297ae83e587fe5a34a8c6b51056d51a42dbdf40df2458
2025/09/02 00:39:21 Loading R1CS from online/test/resources/server.r1cs
2025/09/02 00:39:21 Loading Phase 1 from online/test/resources/server.ph1
2025/09/02 00:39:21 Generating SRS commons form Phase 1
2025/09/02 00:39:21 Storing SRS commons to tiny.srs
2025/09/02 00:39:21 Initializing Phase 2
2025/09/02 00:39:22 Storing Phase 2 to tiny.ph2
2025/09/02 00:39:22 Operation successful
```

#### `contribute`

Contribute randomness to Phase 2. This step is performed by all the participants of the ceremony.

The Coordinator must provide the existing Phase 2 file to the next Contributor. The provided file is either created
in the [`init`](#init) step or in the previous run of the [`contribute`](#contribute) step. After the contribution
is done, the Contributor sends the updated Phase 2 file to the Coordinator for verification (see [`verify`](#verify)).
When the contribution is done, the new Phase 2 file is stored under the same name as the input file, with a timestamp
appended to the name.

- `--phase2` - The existing Phase 2 file created in the `init` step or in the previous run of the `contribute` step.

Example usage:

```shell
$ ./trusted-setup contribute --phase2 tiny.ph2                                                                               
2025/09/02 00:40:17 Contribution to Phase 2:
        Load Phase 2 from:    tiny.ph2
2025/09/02 00:40:17 Loading Phase 2 from tiny.ph2
2025/09/02 00:40:17 Contributing randomness to Phase 2
2025/09/02 00:40:17 Storing Phase 2 to tiny.ph2.20250902004017.451331
2025/09/02 00:40:17 Phase2 file with contributions: tiny.ph2.20250902004017.451331
2025/09/02 00:40:17 Operation successful

$ ./trusted-setup contribute --phase2 tiny.ph2.20250902004017.451331
2025/09/02 00:40:24 Contribution to Phase 2:
        Load Phase 2 from:    tiny.ph2.20250902004017.451331
2025/09/02 00:40:24 Loading Phase 2 from tiny.ph2.20250902004017.451331
2025/09/02 00:40:24 Contributing randomness to Phase 2
2025/09/02 00:40:24 Storing Phase 2 to tiny.ph2.20250902004024.355772
2025/09/02 00:40:24 Phase2 file with contributions: tiny.ph2.20250902004024.355772
2025/09/02 00:40:24 Operation successful
```

#### `verify`

Verify the last randomness contributed to Phase 2. This step is performed by the Coordinator.

This command accepts two Phase 2 files to verify the contributions: previous and next. The previous contribution
is the file that was sent to a Contributor as an input for their contribution process. The next contribution
is the output of that contribution process, that was sent back by the Contributor to the Coordinator.

If the verification is successful, the Coordinator can either:
- send the next contribution file to the next Contributor for further contributions, or
- export the Proving and Verifying Keys (see [`extract-keys`](#extract-keys).)

- `--phase2prev` - A Phase 2 file being an input to the contribution
- `--phase2next` - A Phase 2 file that was contributed to.

Example usage:

```shell
$ ./trusted-setup verify --phase2prev tiny.ph2 --phase2next tiny.ph2.20250902004017.451331
2025/09/02 00:41:17 Verify single Phase 2 contribution:
        Load previous Phase 2 from: tiny.ph2
        Load next Phase 2 from:     tiny.ph2.20250902004017.451331
2025/09/02 00:41:17 Loading Phase 2 from tiny.ph2
2025/09/02 00:41:17 Loading Phase 2 from tiny.ph2.20250902004017.451331
2025/09/02 00:41:17 Verifying the most recent Phase 2 against the previous step
2025/09/02 00:41:17 Operation successful

$ ./trusted-setup verify --phase2prev tiny.ph2.20250902004017.451331 --phase2next tiny.ph2.20250902004024.355772
2025/09/02 00:44:49 Verify single Phase 2 contribution:
        Load previous Phase 2 from: tiny.ph2.20250902004017.451331
        Load next Phase 2 from:     tiny.ph2.20250902004024.355772
2025/09/02 00:44:49 Loading Phase 2 from tiny.ph2.20250902004017.451331
2025/09/02 00:44:49 Loading Phase 2 from tiny.ph2.20250902004024.355772
2025/09/02 00:44:49 Verifying the most recent Phase 2 against the previous step
2025/09/02 00:44:49 Operation successful
```

#### `extract-keys`

Extract the Proving and Verifying Keys. This step is performed by the Coordinator.

This command extracts the Proving and Verifying Keys from the constraint system object, contributions and SRS commons.
The output are binary files containing the keys.

- `--r1cs` - The R1CS file generated from the Gnark circuit the ceremony is held for,
- `--srscommons` - The circuit-independent components of the Groth16 SRS file generated on the [Phase 2 initialization](#init).
- `--beacon` - The beacon value output by the [`init`](#init) command.
- `--phase2` - A list of Phase 2 files to verify the contributions in the order they were created. Contributions are
               verified in pairs, so at least two files must be provided. This DOES NOT INCLUDE the original Phase 2.
               file generated on initialization.
- `--pk` - The output path for the Proving Key file,
- `--vk` - The output path for the Verifying Key file.

Example usage:

```shell
$ ./trusted-setup extract-keys --r1cs online/test/resources/server.r1cs --srscommons tiny.srs --beacon b570d6c3cb53603fc4d297ae83e587fe5a34a8c6b51056d51a42dbdf40df2458 --phase2 tiny.ph2.20250902004017.451331,tiny.ph2.20250902004024.355772 --pk tiny.pk --vk tiny.vk
2025/09/02 00:44:20 Verify multiple Phase 2 contributions:
        Load R1CS from:         online/test/resources/server.r1cs
        Load SRS commons from:  tiny.srs (beacon: b570d6c3cb53603fc4d297ae83e587fe5a34a8c6b51056d51a42dbdf40df2458)
        Load Phase 2 from:      [tiny.ph2.20250902004017.451331 tiny.ph2.20250902004024.355772]
        Store Proving Key to:   tiny.pk
        Store Verifying Key to: tiny.vk
2025/09/02 00:44:20 Loading R1CS from online/test/resources/server.r1cs
2025/09/02 00:44:20 Loading SRS commons from tiny.srs
2025/09/02 00:44:20 Loading Phase 2 from tiny.ph2.20250902004017.451331
2025/09/02 00:44:20 Loading Phase 2 from tiny.ph2.20250902004024.355772
2025/09/02 00:44:20 Verifying all Phase 2 contributions and generating Keys
2025/09/02 00:44:20 Storing Proving Key to tiny.pk
2025/09/02 00:44:20 Storing Verifying Key to tiny.vk
2025/09/02 00:44:20 Operation successful
```
