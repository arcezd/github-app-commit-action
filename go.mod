module github.com/arcezd/github-app-commit-action

go 1.26

replace github.com/arcezd/github-app-commit-action/helper => ./helper

require github.com/arcezd/github-app-commit-action/helper v0.0.0-20240517223547-5b41455a9cac

require github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
