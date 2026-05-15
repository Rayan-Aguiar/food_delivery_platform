module food_delivery_platform/services/auth-service

go 1.25.0

require food_delivery_platform/shared v0.0.0

require github.com/google/uuid v1.6.0 // indirect

require (
	github.com/joho/godotenv v1.5.1
	golang.org/x/crypto v0.51.0
)

replace food_delivery_platform/shared => ../../shared
