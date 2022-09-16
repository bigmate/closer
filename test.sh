# run tests on by one
go test . -race -run New
go test . -race -run Replaces
go test . -race -run Add
go test . -race -run CloseTimeout
go test . -race -run CloseWithoutTimeout
