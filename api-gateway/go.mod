module food_delivery_platform/api-gateway

go 1.25.0

require (
	food_delivery_platform/shared v0.0.0
	github.com/joho/godotenv v1.5.1
)

require github.com/google/uuid v1.6.0 // indirect

replace food_delivery_platform/shared => ../shared
