# grpc-testing

This is the testing and helper tool for grpc server, using [cuelang](https://github.com/cuelang/cue) and [grpcurl](https://github.com/fullstorydev/grpcurl).

notice: This tool is a in progress.

## Usage

### install

```bash
go get -u github.com/ryoya-fujimoto/grpc-testing
```

### Generate test file

First, create a test file using grpc-testing.

```bash
grpc-testing add tests/FirstTest
```

`FirstTest` is the test file name. This command generate cuelang file like this.

tests/FirstTest.cue

```
name: "FirstTest2"
Input: {}
Output: {}
TestCase :: {
	input: Input
	output: Output
}
Test :: {
	name: string
	method: string
	proto?: [...string]
	import_path?: [...string]
	headers?: [string]: string
	input?: Input
	output?: Output
	tests?: [...TestCase]
}
cases: [...Test] & [
	{
		name: ""
		method: ""
		tests: [
			input: {}
			output: {}
		]
	},
]
```

`cases` is a test cases. You can add grpc test case to this list.
The `add` command can specify protobuf files, and when specified, generate cue file is merged protobuf schemas.

```bash
grpc-testing add --proto_path example/app --protofiles example/app/*.proto tests/FirstTest
```

This command generate below cue file.

```
import "github.com/ryoya-fujimoto/grpc-testing/example/app"

name: "FirstTest"
Input: {}
Output: {}
TestCase :: {
	input: Input
	output: Output
}
Test :: {
	name: string
	method: string
	proto?: [...string]
	import_path?: [...string]
	headers?: [string]: string
	input?: Input
	output?: Output
	tests?: [...TestCase]
}
cases: [...Test] & [
	{
		name: ""
		method: ""
		tests: [
			input: {}
			output: {}
		]
	},
]
```

### Test your grpc server

Edit your test case file for testing grpc server, like below (write cases param only).

```
cases: [...Test] & [{
	name:   "GetUser"
	method: "UserService.GetUser"
	input: app.GetUserRequest & {
		id: 5
 	}
 	output: {
		id: "5"
		name: "John Smith"
	}
}]
```

Now, you can request to grpc server using input object.

`grpc-testing run` prints response from server.

```bash
$ grpc-testing run localhost:8080 tests/FirstTest.cue
tests/FirstTest.cue
        test name: GetUser
        method: UserService.GetUser
        output: {
                  "id": "5",
                  "name": "John Smith"
                }
```

`grpc-testing test` compares between response and output parameter.

```bash
$ grpc-testing test localhost:8080 tests/FirstTest.cue
tests/FirstTest.cue
        OK: GetUser
```

### Test your grpc server using protofile, not gRPC reflection API

If your grpc server does not implement gRPC reflection API, you can use protofiles in the same way as the grpcurl `-proto` and `-import-path` options.

add `proto` and `import_path` to test case setting:

```
cases: [...Test] & [{
	name:   "GetUser"
	method: "UserService.GetUser"
	proto: ["./example/app/app.proto"]
	import_path: ["./example/app/"]
	input: app.GetUserRequest & {
		id: 5
	}
	output: {
		id: "5"
		name: "John Smith"
	}
}]
```
