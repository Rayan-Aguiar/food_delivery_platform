package broker

type Config struct {
	URL          string
	ExchangeName string
	ExchangeType string
	ConsumerTag  string
	Prefetch     int
}
