module github.com/arcezd/github-app-commit-action

go 1.22.1

replace github.com/arcezd/github-app-commit-action/helper => ./helper

require github.com/arcezd/github-app-commit-action/helper v0.0.0-00010101000000-000000000000

require github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
