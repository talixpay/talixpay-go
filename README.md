# talixpay-go

Server-side Go SDK for TalixPay.

```go
import (
  "context"
  talixpay "github.com/talixpay/talixpay-go"
)

client := talixpay.New("https://api.talixpay.com", os.Getenv("TALIX_API_KEY"))

session, err := client.CreateCashierSession(ctx, talixpay.CreateCashierSessionInput{
  Mode:   "deposit",
  Amount: 25,
  AllowedOrigins: []string{"https://yoursite.example"},
})
// pass session.SessionToken to @talixpay/sdk in the browser

ok := talixpay.VerifyWebhook(rawBody, webhookSecret, r.Header.Get("X-Signature"))
```

```bash
go get github.com/talixpay/talixpay-go@v0.1.0
```
