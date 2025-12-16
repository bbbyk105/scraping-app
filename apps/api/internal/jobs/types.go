package jobs

const TypeFetchPrices = "fetch_prices"

type FetchPricesPayload struct {
	Source string `json:"source"` // "demo", "public_html", or "all"
}

