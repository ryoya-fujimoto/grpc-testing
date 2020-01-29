# grpc-testing

This is the testing and helper tool for grpc server, using [cuelang](https://github.com/cuelang/cue) and [grpcurl](https://github.com/fullstorydev/grpcurl).

notice: This tool is a in progress.

## Usage

### install

```bash
go get github.com/ryoya-fujimoto/grpc-testing
```

### Generate test file

First, create a test file using grpc-testing.

```bash
grpc-testing add FirstTest 
```

`FirstTest` is the test case name. This command generate cuelang file like this.

```
{
	name: "FirstTest"
	Input: {
	}
	Output: {
	}
	Test :: {
		method: string
		input:  Input
		output: Output
	}
	cases: [...Test] & [{
		method: ""
		input: {
		}
		output: {
		}
	}]
}
```

`cases` is a test cases. You can add grpc test case to this list.
The `add` command can specify protobuf files, and when specified, generate cue file is merged protobuf schemas.

```bash
grpc-testing add --proto_path example/app --protofiles example/app/*.proto FirstTest
```

This command generate below cue file.

```
{
	name: "FirstTest"
	Input: {
	}
	Output: {
	}
	Test :: {
		method: string
		input:  Input
		output: Output
	}
	cases: [...Test] & [{
		method: ""
		input: {
		}
		output: {
		}
	}]
	GetUserRequest: {
		id?: uint64 @protobuf(1)
	}
	CreateUserRequest: {
		name?: string @protobuf(1)
	}
	User: {
		name?: string @protobuf(2)
		id?:   uint64 @protobuf(1)
	}
}
```

### Test your grpc server

Edit your test case file for testing grpc server, like below (write cases param only).

```
cases: [...Test] & [{
  method: "UserService.GetUser"
  input: GetUserRequest & {
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
$ grpc-testing run localhost:8080 FirstTest
output: {
  "id": "5",
  "name": "John Smith"
}
```

`grpc-testing test` compares between response and output parameter.
```bash
$ grpc-testing test localhost:8080 FirstTest
OK: addTest
```
