# ZKP Trusted Setup Ceremony Coordinator

**Warning**
Please note that this tool is under development. Please consider it unusable before the first release.

## Overview
This utility program allows for performing a Trusted Setup Ceremony in a Multi-Party Computation fashion. It is meant
to be used by the Coordinator of the ceremony, as well as by the Contributors. In the end, the Coordinator will obtain
Proving and Verifying Keys, which can be used to generate proofs for the circuit the ceremony was conducted for.

### Online mode

The primary mode of the program. In this mode, the Coordinator runs the ceremony server, which is responsible for
accepting contributions from the Contributors. The Contributors connect to the Coordinator and contribute to the
ceremony.

See help for `server` and `client` commands for details.

### Offline mode

In this mode, the Coordinator and the Contributors run the ceremony locally. The Coordinator initializes the ceremony
and generates the initial Phase 2 file. The Coordinator sends the file to the first Contributor. The Contributor
generates their contribution and sends them to the Coordinator in the form of a Phase 2 file. The Coordinator verifies
the contributions and, if the verification is positive, sends it to the next Contributor.

In this mode, sending Phase 2 files must be performed manually by the Coordinator and Contributors.

At the end of the ceremony, the Coordinator will have a list of accepted contributions. The Coordinator can then
perform the final verification and extract the Proving and the Verifying Keys.

See help for `init`, `contrib`, `verify` and `extract` commands for details.

### Snarkjs powers of tau (ptau) -> Phase 1 conversion

The tool can convert a Snarkjs powers of tau file to a Phase 1 file. This step is performed by the Coordinator before
the initialization of the offline mode ceremony, if the Coordinator has a ptau file that they wish to use in the ceremony.

This step is not necessary if the Coordinator already has a Phase 1 file.

## Constraints

Gnark version used for implementing the circuit the ceremony will be conducted for must match the Gnark version used
in this project. Please consult `[go.mod](./go.mod)` to learn which version of Gnark is used.

Your Gnark project must satisfy the following constraints:
- Supported curve: BN254
- Supported backend: Groth16

## Prerequisites

These are one-time steps that must be done in order to build the program.

Install [Go](https://go.dev/dl/). Any recent version will do. Look into `go.mod` to see the minimum required version.

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
./trusted-setup <command> <options>
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

### Online mode commands

#### `server`

Start a Ceremony server. This step is performed by the Coordinator.

The server is responsible for orchestrating the ceremony, receiving contributions from the participants and, in the end,
generating Proving and Verifying Keys.

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
}
```

Coordination of the ceremony is automatic. No action from the Coordinator is required besides starting the server
and stopping it with CTRL+C at any arbitrary moment. At CTRL+C, the server stops accepting new contributions and starts
key extraction from the existing contributions.

- `--config` - Path to a JSON file containing the server configuration.

#### `client`

Connect to a Ceremony server and provide contributions. This step is performed by the Contributors.

The client is responsible for connecting to the server and providing contributions. The client is configured with
a host and port of the server. Participation in the ceremony is automatic. No action from the Contributor is required
besides starting the client.

- `--host` - The IP address of the server,
- `--port` - The port of the server.

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

#### `contribute`

Contribute randomness to Phase 2. This step is performed by all the participants of the ceremony.

The Coordinator must provide the existing Phase 2 file to the next Contributor. The provided file is either created
in the [`init`](#init) step or in the previous run of the [`contribute`](#contribute) step. After the contribution
is done, the Contributor sends the updated Phase 2 file to the Coordinator for verification (see [`verify`](#verify)).
When the contribution is done, the new Phase 2 file is stored under the same name as the input file, with a timestamp
appended to the name.

- `--phase2` - The existing Phase 2 file created in the `init` step or in the previous run of the `contribute` step.

#### `verify`

Verify the last randomness contributed to Phase 2. This step is performed by the Coordinator.

This command accepts two Phase 2 files to verify the contributions: previous and next. The previous contribution
is the file that was sent to a Contributor as an input for their contribution process. The next contribution
is the output of that contribution process, that was sent back by the Contributor to the Coordinator.

If the verification is successful, the Coordinator can either:
- send the next contribution file to the next Contributor for further contributions, or
- export the Proving and Verifying Keys (see [`keys`](#keys).)

- `--phase2prev` - A Phase 2 file being an input to the contribution
- `--phase2next` - A Phase 2 file that was contributed to.

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
