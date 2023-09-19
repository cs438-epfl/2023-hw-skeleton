export GLOG = warn
export BINLOG = warn
export HTTPLOG = warn
export GORACE = halt_on_error=1

all: test lint vet

test: test_hw0 test_hw1 test_hw2 test_hw3

test_hw0: test_unit_hw0 test_int_hw0
test_hw1: test_unit_hw1 test_int_hw1
test_hw2: test_unit_hw2 test_int_hw2
test_hw3: test_unit_hw3 test_int_hw3

test_unit_hw0:
	go test -timeout 2m -v -race -run Test_HW0 ./peer/tests/unit

test_unit_hw1:
	go test -timeout 2m -v -race -run Test_HW1 ./peer/tests/unit

test_unit_hw2:
	go test -timeout 2m -v -race -run Test_HW2 ./peer/tests/unit

test_unit_hw3:
	go test -timeout 2m -v -race -run Test_HW3 ./peer/tests/unit

test_int_hw0:
	go test -timeout 5m -v -race -run Test_HW0 ./peer/tests/integration

test_int_hw1:
	go test -timeout 5m -v -race -run Test_HW1 ./peer/tests/integration

test_int_hw2:
	go test -timeout 5m -v -race -run Test_HW2 ./peer/tests/integration

test_int_hw3:
	go test -timeout 5m -v -race -run Test_HW3 ./peer/tests/integration

lint:
	# Coding style static check.
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.0
	@go mod tidy
	golangci-lint run

vet:
	go vet ./...

