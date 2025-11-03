module github.com/mikeschinkel/go-fsfix

go 1.25.3

replace (
	github.com/mikeschinkel/go-dt => ../go-dt
	github.com/mikeschinkel/go-dt/de => ../go-dt/de
)

require github.com/mikeschinkel/go-dt v0.0.0-20251103071339-3bc3e3719d7e

require github.com/mikeschinkel/go-dt/de v0.0.0-00010101000000-000000000000 // indirect
