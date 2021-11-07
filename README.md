# hedera-setup-go
Command line tool to perform basic setup on the hedera network.

## Running locally
Using zenv setup environment `source zenv hedera-sandbox`
### create account
```bash
go run cmd/main.go create-account -network testnet -operator-account $ACCOUNT_ID -operator-private-key $PRIVATE_KEY
```

### create topic
haven't tested yet
