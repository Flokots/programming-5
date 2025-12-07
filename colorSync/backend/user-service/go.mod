module github.com/Flokots/programming-5/colorSync/user-service

go 1.25.4

require (
	github.com/Flokots/programming-5/colorSync/shared v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.45.0
)

require github.com/golang-jwt/jwt/v5 v5.3.0 // indirect

replace github.com/Flokots/programming-5/colorSync/shared => ../shared
