# run tests on by one
go test . -race -run Add
go test . -race -run Options
go test . -race -run Reconfigure
go test . -race -run CloseTimeout
go test . -race -run CloseWithoutTimeout
